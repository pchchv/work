package work

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

func newTestPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxActive:   10,
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		Wait: true,
	}
}

func cleanKeyspace(namespace string, pool *redis.Pool) {
	conn := pool.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", namespace+"*"))
	if err != nil {
		panic("could not get keys: " + err.Error())
	}

	for _, k := range keys {
		if _, err := conn.Do("DEL", k); err != nil {
			panic("could not del: " + err.Error())
		}
	}
}

func zsetSize(pool *redis.Pool, key string) int64 {
	conn := pool.Get()
	defer conn.Close()

	v, err := redis.Int64(conn.Do("ZCARD", key))
	if err != nil {
		panic("could not get ZSET size: " + err.Error())
	}
	return v
}

func listSize(pool *redis.Pool, key string) int64 {
	conn := pool.Get()
	defer conn.Close()

	v, err := redis.Int64(conn.Do("LLEN", key))
	if err != nil {
		panic("could not get list length: " + err.Error())
	}
	return v
}
