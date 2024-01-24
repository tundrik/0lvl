package repository

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"0lvl/config"
	"0lvl/internal/inspector"
	"0lvl/pkg/cache"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

const (
	// maxCacheBytes - максимальный размер кеша (FIFO).
	maxCacheBytes = 1024 * 1024 * 32

	// лимит каждого запроса при заполнении кеша при старте
	selectLimit = 1024

	urlOrder = "http://localhost:8000/order/"
)

type Repo struct {
	db    *pgxpool.Pool
	cache *cache.Cache
	log   zerolog.Logger
}

// Инициализирует репозиторий.
// 1) Пул подключений к базе данных
// 2) Кеш
func New(ctx context.Context, cfg config.Config, log zerolog.Logger) (*Repo, error) {
	db, err := pgxpool.New(ctx, cfg.PgString)
	if err != nil {
		return nil, err
	}

	cache, err := cache.New(maxCacheBytes)
	if err != nil {
		return nil, err
	}

	repo := &Repo{
		db:    db,
		cache: cache,
		log:   log,
	}

	return repo, nil
}

// Возвращает заказ из кеша, если в кеше нет то из базы данных
func (r *Repo) Order(uid string) ([]byte, error) {
	var b []byte

	b, ok := r.cache.HasGet(b, []byte(uid))
	if ok {
		return b, nil
	}

	const sql = `SELECT entity FROM trade WHERE pk = $1;`
	err := r.db.QueryRow(context.Background(), sql, uid).Scan(&b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Собирает sql в пакет и отправляет в базу
// успешно сохраненные сохраняет в кеш по одному
// при ошибке INSERT в кеш непоподает, а в box добавляется ошибка pg.
func (r *Repo) SaveOrderBatch(batch []*inspector.OrderBox) []*inspector.OrderBox {
	defer timer(r.log)(len(batch))

	pgBatch := &pgx.Batch{}
	const sql = `INSERT INTO trade (pk, rang, entity) VALUES ($1, $2, $3);`

	for _, box := range batch {
		pgBatch.Queue(sql, box.Uid, box.Rang, box.Data)
	}

	results := r.db.SendBatch(context.Background(), pgBatch)
	defer results.Close()

	for _, box := range batch {
		_, err := results.Exec()

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				box.Err = pgErr
			}
			continue
		}
		r.cache.Set([]byte(box.Uid), box.Data)
	}

	return batch
}

// Возвращает список ссылок заказов на детальный просмотр.
func (r *Repo) OrdersLink(count int) []byte {
	const sql = `SELECT pk, rang FROM trade ORDER BY rang DESC LIMIT $1;`

	rows, err := r.db.Query(context.Background(), sql, count)
	if err != nil {
		r.log.Err(err).Msg("")
	}
	defer rows.Close()

	entities := make([]OrderLink, 0, count)

	for rows.Next() {
		rowValues := rows.RawValues()
		pk := string(rowValues[0])

		var sb strings.Builder
		sb.WriteString(urlOrder)
		sb.WriteString(pk)

		entity := OrderLink{
			Uid:  pk,
			Link: sb.String(),
		}
		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		r.log.Err(err).Msg("db error")
	}

	b, _ := json.Marshal(entities)
	return b
}

// Возвращает информацию по кешу и базе данных.
func (r *Repo) Metrica() []byte {
	var m Monitor
	r.cache.UpdateStats(&m.Cache)

	const sql = `SELECT count(pk) FROM trade;`
	err := r.db.QueryRow(context.Background(), sql).Scan(&m.DatabaseOrderCount)
	if err != nil {
		r.log.Err(err).Msg("db error")
	}

	mb, _ := json.Marshal(m)
	return mb
}

// Рекурсивно заполняет кеш заказами из базы данных.
// Пока в одном из баскетов кеша не потребуется удалять записи
// Это заполняет кеш на 50%, хз как его заполнить чтобы более 
// старые записи не затерли свежие.
// В данной реализации кеша с баскетами
func (r *Repo) СacheWarmUp() {
	r.cacheWarmUpChank(selectLimit, 0)
}

func (r *Repo) cacheWarmUpChank(limit int, cursor uint64) {
	var sql strings.Builder
	namedArgs := make(map[string]interface{})
	namedArgs["limit"] = limit

	sql.WriteString("SELECT pk, rang, entity FROM trade")

	if cursor != 0 {
		sql.WriteString(" WHERE rang < @cursor")
		namedArgs["cursor"] = cursor
	}

	sql.WriteString(" ORDER BY rang DESC LIMIT @limit;")

	rows, err := r.db.Query(context.Background(), sql.String(), pgx.NamedArgs(namedArgs))
	if err != nil {
		r.log.Err(err).Msg("db error")
	}
	defer rows.Close()

	for rows.Next() {
		rowValues := rows.RawValues()
		hasNeedClean := r.cache.StopSet(rowValues[0], rowValues[2])

		if hasNeedClean {
			return
		}

		cursor = binary.BigEndian.Uint64(rowValues[1])
	}

	if err := rows.Err(); err != nil {
		r.log.Err(err).Msg("db error")
	}

	r.cacheWarmUpChank(selectLimit, cursor)
}

func timer(logger zerolog.Logger) func(c int) {
	start := time.Now()
	return func(c int) {
		logger.Info().Int("count orders", c).Int("milliseconds", int(time.Since(start).Milliseconds())).Msg("done save")
	}
}
