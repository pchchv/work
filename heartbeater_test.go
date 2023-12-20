package work

import "github.com/gomodule/redigo/redis"

func redisInSet(pool *redis.Pool, key, member string) bool {
	conn := pool.Get()
	defer conn.Close()

	v, err := redis.Bool(conn.Do("SISMEMBER", key, member))
	if err != nil {
		panic("could not delete retry/dead queue: " + err.Error())
	}
	return v
}
