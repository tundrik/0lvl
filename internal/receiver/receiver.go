package receiver

import (
	"time"

	"0lvl/config"
	"0lvl/internal/inspector"
	"0lvl/internal/repository"

	stan "github.com/nats-io/stan.go"
	"github.com/rs/zerolog"
)


const (
	// Receiver будет накапливать заказы до bufLimit
	// и отправлять на сохранение пакетом для снижения нагрузки
	// на Postgres.
	bufLimit = 128

	// если трафик низкий не ждем накопления а сохраняем по истечению deadline.
	deadline = 500 * time.Millisecond

	// maxInflight — опция, позволяющая установить максимальное количество сообщений, которые будет отправлять кластер.
	// без подтверждения.
	maxInflight = 128

	// повторно заказы прилетят из stan если их не подтвердить за natsAskWait
	// В редких случаях при получении повторных заказов
	// receiver пологается на Postgres ограничение уникальности
	natsAskWait = deadline * 3
)


type Receiver struct {
	id   string
	conn stan.Conn

	cfg  config.Config
	repo *repository.Repo
	log  zerolog.Logger
}

func New(id string, repo *repository.Repo, cfg config.Config, log zerolog.Logger) (stan.Conn, error) {
	conn, err := stan.Connect(cfg.StanClusterId, cfg.StanClientId+id)
	if err != nil {
		return nil, err
	}

	r := &Receiver{
		id:   id,
		conn: conn,
		cfg:  cfg,
		repo: repo,
		log:  log,
	}

	buf := make(chan inspector.MsgBox, bufLimit)

	err = r.subscribe(buf)
	if err != nil {
		return nil, err
	}

	go r.cumulate(buf)

	return conn, nil
}

func (r *Receiver) subscribe(buf chan<- inspector.MsgBox) error {
	ins := inspector.New()

	accept := func(m *stan.Msg) {
		newBox := inspector.MsgBox{
			Msg:  m,
			Data: m.Data,
		}

		box := ins.Audit(newBox)

		if box.Err != nil {
			r.log.Error().Err(box.Err).Msg("audit error")
			//TODO:подтвердить с ошибкой
			m.Ack()
			return
		}
		buf <- box
	}

	_, err := r.conn.QueueSubscribe(
		r.cfg.StanSubject,
		r.cfg.StanQueue,
		accept,
		stan.DurableName(r.cfg.StanClientId+r.id),
		stan.SetManualAckMode(),
		stan.AckWait(natsAskWait),
		stan.MaxInflight(maxInflight),
	)

	//QueueSubscribe закроется неявно при закрытии stan.Conn
	return err
}

func (r *Receiver) cumulate(buf <-chan inspector.MsgBox) {
	batch := make([]*inspector.MsgBox, 0, bufLimit)

	flush := func() {
		results := r.repo.SaveOrderBatch(batch, r.id)
		for _, box := range results {
			if box.Err != nil {
				//TODO:подтвердить с ошибкой
				r.log.Error().Err(box.Err).Msg("save error")
				box.Msg.Ack()
			}
			box.Msg.Ack()
		}

		batch = batch[:0]
	}

	ticker := time.NewTicker(deadline)

	for {
		select {
		case <-ticker.C:
			if len(batch) > 0 {
				flush()
			}

		case box := <-buf:
			batch = append(batch, &box)

			if len(batch) == bufLimit {
				flush()
				ticker.Reset(deadline)
			}

		}
	}
}
