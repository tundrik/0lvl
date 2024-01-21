package repository

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"strings"

	"0lvl/config"
	"0lvl/internal/inspector"
	"0lvl/pkg/cache"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)


const (
	// maxCacheBytes максимальный размер кеша (FIFO).
	maxCacheBytes = 1024 * 1024 * 32

	// entryByteSize примерный размер кешированного ордера в байтах.
	// если это будет меньше чем по факту то запрос в db при старте будет с большим лимитом
	// чем влезет в кеш
	//
	// при двух item в ордере это 1000 - 1300
	entryByteSize = 1280

	// initCacheCount количество ордеров из db для прогрева кеша при старте.
	initCacheCount = maxCacheBytes / entryByteSize

	// поскольку initCacheCount может быть довольно большим
	// то поделим на чанки maxSelectLimit
	maxSelectLimit = 20000
)


type Repo struct {
	db    *pgxpool.Pool
	cache *cache.Cache
	log   zerolog.Logger
}

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
	go repo.СacheWarmUp()
	//не ждем пока Сache заполнится идем дальше

	return repo, nil
}

// возвращает заказ из кеша, если в кеше нет то из базы
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

// собирает sql в пакет и отправляет в базу
// успешно сохраненные сохраняет в кеш по одному
//
// при ошибке INSERT в кеш непоподает, а в box добавляется ошибка pg
func (r *Repo) SaveOrderBatch(batch []*inspector.MsgBox, id string) []*inspector.MsgBox {
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

	r.log.Debug().Int("len", len(batch)).Str("receiver id", id).Msg("")
	return batch
}

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

		entity := OrderLink{
			Uid:  pk,
			Link: "http://localhost:8000/order/" + pk,
		}
		entities = append(entities, entity)
	}
	if err := rows.Err(); err != nil {
		r.log.Err(err).Msg("db error")
	}

	b, _ := json.Marshal(entities)
	return b
}

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

func (r *Repo) СacheWarmUp() {
	var cursor uint64

	for i := 0; i < initCacheCount/maxSelectLimit; i++ {
		cursor = r.СacheWarmUpChank(maxSelectLimit, cursor)
	}

	mod := initCacheCount % maxSelectLimit
	if mod > 0 {
		r.СacheWarmUpChank(mod, cursor)
	}
	r.log.Info().Msg("done cache warm up")
}

func (r *Repo) СacheWarmUpChank(limit int, cursor uint64) uint64 {
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
		r.cache.Set(rowValues[0], rowValues[2])

		cursor = binary.BigEndian.Uint64(rowValues[1])
	}

	if err := rows.Err(); err != nil {
		r.log.Err(err).Msg("db error")
	}

	return cursor
}
