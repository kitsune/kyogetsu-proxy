//Copyright Dylan Enloe 2017

package kyogetsu

import (
  "github.com/mediocregopher/radix.v2/pool"
  "fmt"
  "net/http"
)

//A CookieCache that uses Redis as it's backend store.
//It stores the cookie data in a Redis HashMap under
//the key <namespace>.<id>
type RedisCache struct {
  pool *pool.Pool
  namespace string
}

//SetCookie extracts the name and value from an
//*http.Cookie and stores in the id's hash map
func (r RedisCache) SetCookie(id string, c *http.Cookie) error {
  k := r.namespacedId(id)
  return r.pool.Cmd("HSET", k, c.Name, c.Value).Err
}

//SetCookies stores the name and value data for
//an array of *http.Cookies in the id's hash map
func (r RedisCache) SetCookies(id string, c []*http.Cookie) error {
  k := r.namespacedId(id)
  m := map[string]string{}
  for _, v := range c {
    m[v.Name] = v.Value
  }
  fmt.Println("Set Cookies:")
  fmt.Println(m)
  s, err := r.pool.Cmd("HMSET", k, m).Str()
  fmt.Println(s)
  return err
  //return r.pool.Cmd("HMSET", k, val).Err
}

//GetCookie gets a cookie stored in Redis
func (r RedisCache) GetCookie(id string, key string) (*http.Cookie, error) {
  k := r.namespacedId(id)
  v, err := r.pool.Cmd("HGET", k, key).Str()
  if err != nil {
    return nil, err
  }
  return &http.Cookie{Name: k, Value: v}, nil
}

//GetCookies gets all cookies stored in Redis for
//a given Id
func (r RedisCache) GetCookies(id string) ([]*http.Cookie, error) {
  k := r.namespacedId(id)
  m, err := r.pool.Cmd("HGETALL", k).Map()
  if err != nil {
    return nil, err
  }
  c := make([]*http.Cookie, 0, len(m))
  for k, v := range m {
    c = append(c, &http.Cookie{Name: k, Value: v})
  }
  return c, nil
}

//ChangeCookiesId uses the rename command to change the
//id that the cookie data is returned under
func (r RedisCache) ChangeCookiesId(old_id string, new_id string) error {
  old_id = r.namespacedId(old_id)
  new_id = r.namespacedId(new_id)

  s, err := r.pool.Cmd("RENAME", old_id, new_id).Str()
  fmt.Println(s)
  return err
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