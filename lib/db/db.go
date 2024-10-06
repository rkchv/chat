package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Client клиент БД
type Client interface {
	DB() DB
	Close() error
}

// Handler - функция, которая выполняется в транзакции
type Handler func(ctx context.Context) error

// QueryLogger функция вызываемая при каждом выполнении запроса, например для логирования/трассировки
type QueryLogger func(ctx context.Context, q Query, args ...interface{}) LogFlush

// LogFlush объект, чей метод будет вызван по окончании запроса (на defer)
type LogFlush interface {
	Flush()
}

// DummyFlush объект-заглушка для сброса логов запросов (если не задан свой QueryLogger)
type DummyFlush struct{}

func (f DummyFlush) Flush() {}

// Transactor интерфейс для работы с транзакциями
type Transactor interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	ReadCommitted(ctx context.Context, f Handler) error
}

// Query обертка над запросом, хранящая имя запроса и сам запрос
type Query struct {
	Name     string
	QueryRaw string
}

// QueryExecer интерфейс для работы с обычными запросами
type QueryExecer interface {
	Exec(ctx context.Context, q Query, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, q Query, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, q Query, args ...interface{}) pgx.Row
	SetQueryLogger(logger QueryLogger)
}

// Pinger интерфейс для проверки соединения с БД
type Pinger interface {
	Ping(ctx context.Context) error
}

// DB интерфейс для работы с БД
type DB interface {
	QueryExecer
	Transactor
	Pinger
	Close()
}
