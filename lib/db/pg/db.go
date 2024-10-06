package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

  "github.com/rkchv/chat/lib/db"
)

var _ db.DB = (*pg)(nil)

type key string

const (
	// TxCtxKey Ключ для сохранения транзакции в контексте
	TxCtxKey key = "tx"
)

type pg struct {
	pool    *pgxpool.Pool
	logFunc db.QueryLogger
}

// NewDB Новый экземпляр обертки клиента к pg
func NewDB(dbc *pgxpool.Pool) db.DB {
	return &pg{pool: dbc}
}

func (p *pg) SetQueryLogger(logger db.QueryLogger) {
	p.logFunc = logger
}

func (p *pg) Exec(ctx context.Context, q db.Query, args ...interface{}) (pgconn.CommandTag, error) {
	deferFlush := p.logQuery(ctx, q, args...)
	defer deferFlush.Flush()

	tx, ok := ctx.Value(TxCtxKey).(pgx.Tx)
	if ok {
		return tx.Exec(ctx, q.QueryRaw, args...)
	}

	return p.pool.Exec(ctx, q.QueryRaw, args...)
}

func (p *pg) Query(ctx context.Context, q db.Query, args ...interface{}) (pgx.Rows, error) {
	deferFlush := p.logQuery(ctx, q, args...)
	defer deferFlush.Flush()

	tx, ok := ctx.Value(TxCtxKey).(pgx.Tx)
	if ok {
		return tx.Query(ctx, q.QueryRaw, args...)
	}

	return p.pool.Query(ctx, q.QueryRaw, args...)
}

func (p *pg) QueryRow(ctx context.Context, q db.Query, args ...interface{}) pgx.Row {
	deferFlush := p.logQuery(ctx, q, args...)
	defer deferFlush.Flush()

	tx, ok := ctx.Value(TxCtxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, q.QueryRaw, args...)
	}

	return p.pool.QueryRow(ctx, q.QueryRaw, args...)
}

func (p *pg) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *pg) Close() {
	p.pool.Close()
}

func (p *pg) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	// Если это вложенная транзакция, пропускаем инициацию новой транзакции и выполняем обработчик.
	tx, ok := ctx.Value(TxCtxKey).(pgx.Tx)
	if ok {
		return tx, nil
	}

	// Стартуем новую транзакцию.
	return p.pool.BeginTx(ctx, txOptions)
}

func (p *pg) ReadCommitted(ctx context.Context, f db.Handler) error {
	txOpts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
	return p.transaction(ctx, txOpts, f)
}

// transaction основная функция, которая выполняет указанный пользователем обработчик в транзакции
func (p *pg) transaction(ctx context.Context, opts pgx.TxOptions, fn db.Handler) (err error) {
	// Стартуем новую транзакцию.
	tx, err := p.BeginTx(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "can't begin transaction")
	}

	// Кладем транзакцию в контекст.
	ctx = MakeContextTx(ctx, tx)

	// Настраиваем функцию отсрочки для отката или коммита транзакции.
	defer func() {
		// восстанавливаемся после паники
		if r := recover(); r != nil {
			err = errors.Errorf("panic recovered: %v", r)
		}

		// откатываем транзакцию, если произошла ошибка
		if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				err = errors.Wrapf(err, "errRollback: %v", errRollback)
			}

			return
		}

		// если ошибок не было, коммитим транзакцию
		if nil == err {
			err = tx.Commit(ctx)
			if err != nil {
				err = errors.Wrap(err, "tx commit failed")
			}
		}
	}()

	if err = fn(ctx); err != nil {
		err = errors.Wrap(err, "failed executing code inside transaction")
	}

	return err
}

func (p *pg) logQuery(ctx context.Context, q db.Query, args ...interface{}) db.LogFlush {
	if p.logFunc != nil {
		flush := p.logFunc(ctx, q, args...)
		return flush
	}

	return db.DummyFlush{}
}

// MakeContextTx добавляет транзакцию в контекст, чтобы последующие вызовы транзакций брали ее
func MakeContextTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, TxCtxKey, tx)
}
