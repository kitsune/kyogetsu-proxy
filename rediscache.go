package kyogetsu

import (
  "github.com/mediocregopher/radix.v2/pool"
  "fmt"
)

type RedisCache struct {
  pool *pool.Pool
  namespace string
}

func (r RedisCache) SetCookie(id string, key string, val string) error {
  k := r.namespacedId(id)
  return r.pool.Cmd("HSET", k, key, val).Err
}

func (r RedisCache) SetCookies(id string, val map[string]string) error {
  k := r.namespacedId(id)
  fmt.Println(val)
  s, err := r.pool.Cmd("HMSET", k, val).Str()
  fmt.Println(s)
  return err
  //return r.pool.Cmd("HMSET", k, val).Err
}

func (r RedisCache) GetCookie(id string, key string) (string, error) {
  k := r.namespacedId(id)
  return r.pool.Cmd("HGET", k, key).Str()
}

func (r RedisCache) GetCookies(id string) (map[string]string, error) {
  k := r.namespacedId(id)
  return r.pool.Cmd("HGETALL", k).Map()
}

func (r RedisCache) namespacedId(id string) string {
  k := r.namespace + "." + id
  fmt.Println(k)
  return k
  //return r.namespace + "." + id
}

// Creates a new RedisCache with only a single connection.
// If more connections are needed they will be created on
// the fly.  This is still a redis pool
func NewRedisCache(addr string) *RedisCache {
  return NewRedisCachePool(addr, 1)
}

func NewRedisCachePool(addr string, size int) *RedisCache {
  p, _ := pool.New("tcp", addr, size)
  return &RedisCache{pool: p, namespace: "kyogetsu"}
}