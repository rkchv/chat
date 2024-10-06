package redis

import (
	"context"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"

	def "github.com/rkchv/chat/lib/redis"
)

var _ def.Client = (*client)(nil)

type handler func(ctx context.Context, conn redis.Conn) error

type client struct {
	conn *redis.Pool
}

// NewClient Новый экземпляр клиента к редис
func NewClient(pool *redis.Pool) def.Client {
	return &client{conn: pool}
}

// Set выполняет команду Set
func (c *client) Set(ctx context.Context, key string, value interface{}) error {
	err := c.exec(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("SET", redis.Args{key}.Add(value)...)

		return err
	})

	return err
}

func (c *client) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := c.exec(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("SETEX", redis.Args{key}.Add(value).Add(int(expiration.Seconds()))...)

		return err
	})

	return err
}

func (c *client) Get(ctx context.Context, key string) (string, error) {
	var res interface{}
	err := c.exec(ctx, func(ctx context.Context, conn redis.Conn) error {
		var errCmd error
		res, errCmd = conn.Do("GET", key)

		return errCmd
	})
	if err != nil {
		return "", err
	}

	return res.(string), nil
}

func (c *client) Del(ctx context.Context, key string) error {
	err := c.exec(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("DEL", key)

		return err
	})

	return err
}

func (c *client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := c.exec(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("EXPIRE", key, int(expiration.Seconds()))

		return err
	})

	return err
}

func (c *client) Exist(ctx context.Context, key string) (bool, error) {
	var rs bool
	err := c.exec(ctx, func(ctx context.Context, conn redis.Conn) error {
		var errCmd error
		rs, errCmd = redis.Bool(conn.Do("EXISTS", key))

		return errCmd
	})

	if err != nil {
		return false, err
	}

	return rs, nil
}

func (c *client) HSet(ctx context.Context, key string, field string, value string) error {
	err := c.exec(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("HSET", key, field, value)

		return err
	})

	return err
}

func (c *client) HSetMap(ctx context.Context, key string, values interface{}) error {
	err := c.exec(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("HSET", redis.Args{key}.AddFlat(values)...)

		return err
	})

	return err
}

func (c *client) HGet(ctx context.Context, key string, field string) (string, error) {
	var res interface{}
	err := c.exec(ctx, func(ctx context.Context, conn redis.Conn) error {
		var errCmd error
		res, errCmd = redis.Values(conn.Do("HGET", redis.Args{key}.Add(field)...))

		return errCmd
	})
	if err != nil {
		return "", err
	}

	return res.(string), nil
}

func (c *client) HDel(ctx context.Context, key string, field string) error {
	err := c.exec(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("HDEL", key, field)

		return err
	})

	return err
}

func (c *client) HGetAll(ctx context.Context, key string, dest interface{}) error {
	var vals []interface{}
	err := c.exec(ctx, func(ctx context.Context, conn redis.Conn) error {
		var errCmd error
		vals, errCmd = redis.Values(conn.Do("HGETALL", key))
		if errCmd != nil {
			return errCmd
		}

		errCmd = redis.ScanStruct(vals, dest)
		if errCmd != nil {
			return errCmd
		}

		return nil
	})

	return err
}

func (c *client) exec(ctx context.Context, handlerFunc handler) error {
	conn := c.conn.Get()
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Printf("redis close conn err: %v", err)
		}
	}()

	return handlerFunc(ctx, conn)
}
