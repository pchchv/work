package work

import "github.com/gomodule/redigo/redis"

func readHash(pool *redis.Pool, key string) map[string]string {
	m := make(map[string]string)

	conn := pool.Get()
	defer conn.Close()

	v, err := redis.Strings(conn.Do("HGETALL", key))
	if err != nil {
		panic("could not delete retry/dead queue: " + err.Error())
	}

	for i, l := 0, len(v); i < l; i += 2 {
		m[v[i]] = v[i+1]
	}
	return m
}
