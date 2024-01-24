package receiver

import (
	"runtime"
	"time"

	"0lvl/config"
	"0lvl/internal/inspector"
	"0lvl/internal/repository"

	stan "github.com/nats-io/stan.go"
	"github.com/rs/zerolog"
)

/* Stan Options */
const (
	// maxInflight — максимальное количество сообщений,
	// которые будет отправлять кластер без подтверждения.
	defaultMaxInflight = 1024

	// Повторно заказы прилетят из stan если их не подтвердить за defaultAckWait
	// В редких случаях при получении повторных заказов
	// Receiver пологается на Postgres ограничение уникальности.
	defaultAckWait = 5 * time.Minute
)

/* Receiver Options */
const (
	// Количество накопителей на каждого подписчика.
	defaultCumCount = 2  

	// defaultMaxSize — максимальный размер пакета заказов отправленных в базу данных.
	defaultMaxSize = 512

	// если трафик низкий пакет отправится по дедлайну не дожидаясь максимального размера.
	defaultDeadline = 256 // ms
)

type Receiver struct {
	conn stan.Conn

	cfg  config.Config
	repo *repository.Repo
	log  zerolog.Logger
}

// Инициализирует ресивер.
// Создает соединение Stan.
func New(repo *repository.Repo, cfg config.Config, log zerolog.Logger) (*Receiver, error) {
	// Этот обратный вызов будет вызван, если клиент окончательно потеряет
	// контакт с сервером (или другой клиент заменяет его во время пребывания conn.Close()).
	connectionLost := func(_ stan.Conn, reason error) {
		log.Error().Err(reason).Msg("stan callback error")
	}
	opt := stan.SetConnectionLostHandler(connectionLost)

	conn, err := stan.Connect(cfg.StanClusterId, cfg.StanClientId, opt)
	if err != nil {
		return nil, err
	}

	rec := &Receiver{
		conn: conn,
		cfg:  cfg,
		repo: repo,
		log:  log,
	}

	return rec, nil
}

// Закрывает соединение stan.Conn.
func (r *Receiver) Close() {
	r.conn.Close()
}

// Запускает подписчиков.
// Каждый подписчик передает канал
// нескольким накопителям.
func (r *Receiver) Run() error {
	// Количество подписчиков.
	subCount := runtime.NumCPU() / 2
	// Количество накопителей на каждого подписчика.
	cumCount := defaultCumCount

	for i := 0; i < subCount; i++ {
		ch, err := r.subscriber()
		if err != nil {
			return err
		}
		r.massCumulate(ch, cumCount)
	}
	return nil
}

// Запускает накопителей с разным таймингом обращений в базу данных.
func (r *Receiver) massCumulate(ch <-chan inspector.OrderBox, cumCount int) {
	size := defaultMaxSize
	deadline := defaultDeadline

	for i := 0; i < cumCount; i++ {
		go r.cumulative(ch, size, deadline)
        // Это немного раскидывает тайминг запросов в базу данных
		// но только в рамках одного подписчика.
		size = size - (size*32)/100
		deadline = deadline - (deadline*16)/100
	}
}

// Подписчик создает канал и подписывается.
// На каждый обратный вызов проверяет данные
// и отправляет по каналу
// который читают несколько накопителей - cumulative
func (r *Receiver) subscriber() (<-chan inspector.OrderBox, error) {
	ch := make(chan inspector.OrderBox)
	ins := inspector.New()

	accept := func(m *stan.Msg) {
		newBox := inspector.OrderBox{
			Msg:  m,
			Data: m.Data,
		}

		box := ins.Audit(newBox)
		if box.Err != nil {
			r.log.Warn().Err(box.Err).Msg("inspector audit error")
			//TODO:подтвердить с ошибкой
			//nats Term()
			m.Ack()
			return
		}
		ch <- box
	}

	_, err := r.conn.QueueSubscribe(
		r.cfg.StanSubject,
		r.cfg.StanQueue,
		accept,
		stan.DurableName("service orders"),
		stan.SetManualAckMode(),
		stan.AckWait(defaultAckWait),
		stan.MaxInflight(defaultMaxInflight),
	)

	//QueueSubscribe закроется неявно при закрытии stan.Conn
	return ch, err
}

// Накопитель принимает проверенные данные,
// при накоплении до лимита или по дедлайну сливает в базу данных.
// Блокируется в ожидании результатов от базы данных.
// Пока одна ждет другая накапливает.
// При низком трафике по большей части они читают по очереди
// (g 1) append 1
// (g 2) append 1
// (g 1) append 1
// (g 2) append 1
// (g 1) flush 2
// (g 2) flush 2
//
// Только при большом трафике работает как ожидалось 
// (g 1) flush 80
// (g 2) append 1
// (g 2) append 1
// (g 2) append ...
// (g 2) flush 126
// (g 1) append 1
// (g 1) append 1
//
func (r *Receiver) cumulative(ch <-chan inspector.OrderBox, size int, deadline int) {
	batch := make([]*inspector.OrderBox, 0, size)
	
	flush := func() {
		results := r.repo.SaveOrderBatch(batch)

		for _, box := range results {
			if box.Err != nil {
				r.log.Error().Err(box.Err).Str("order uid", box.Uid).Msg("save error")
				//TODO:подтвердить с ошибкой
				//nats Term()
				box.Msg.Ack()
			}
			box.Msg.Ack()
		}

		batch = batch[:0]
			}

	d := time.Duration(deadline)
	ticker := time.NewTicker(d * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			if len(batch) > 0 {
				flush()
			}

		case box := <-ch:
			batch = append(batch, &box)
			
			if len(batch) == size {
				flush()
				ticker.Reset(d * time.Millisecond)
			}
		}
	}
}
