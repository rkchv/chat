package redis

import (
	"context"
	"time"
)

// Client Клиент-обертка к редису
type Client interface {
	Set(ctx context.Context, key string, value interface{}) error
	SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Exist(ctx context.Context, key string) (bool, error)
	HSet(ctx context.Context, key string, field string, value string) error
	HSetMap(ctx context.Context, key string, values interface{}) error
	HGet(ctx context.Context, key string, field string) (string, error)
	HDel(ctx context.Context, key string, field string) error
	HGetAll(ctx context.Context, key string, dest interface{}) error
}
