/* Copyright Dylan Enloe 2017
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package kyogetsu

import (
  "github.com/mediocregopher/radix.v2/pool"
  "testing"
  "net/http"
  )

type cookieData struct {
  Name string
  Value string
}

//get a redis connection
func getRedisConn(rc *RedisCache) *pool.Pool {
  e := rc.pool.Cmd("FLUSHALL").Err
  if e != nil {
    panic(e)
  }
  return rc.pool
}

//Get a RedisCache, provides a single place to change
//the host address
func getRedisCache() *RedisCache {
  return NewRedisCache("127.0.0.1:6379")
}

//Close and clean up the redis connection
func closeRedisConn(r *pool.Pool) {
  e := r.Cmd("FLUSHALL").Err
  if e != nil {
    panic(e)
  }
}

//Get the redis path for a given session if
func getIdPath(id string) string{
  return "kyogetsu." + id
}

func TestSetCookie(t *testing.T) {
  var tests = []struct {
      Id string
      cd cookieData
    }{
      {"bill", cookieData{"type", "test"}},
      {"bob", cookieData{"lastname", ""}},
      {"", cookieData{"name", "sessionless"}},
    }
  for _, test := range tests {
    rc := getRedisCache()
    r := getRedisConn(rc)
    defer closeRedisConn(r)
    c := http.Cookie{Name: test.cd.Name, Value: test.cd.Value}

    err := rc.SetCookie(test.Id, &c)
    if err != nil {
      t.Errorf("Got Error: %s", err)
      return
    }

    var v string
    v, err = r.Cmd("HGET", getIdPath(test.Id), test.cd.Name).Str()
    if err != nil {
      t.Errorf("Got Error: %s", err)
      return
    }

    if test.cd.Value != v {
      t.Errorf("Expected: %s, Got: %s", test.cd.Value, v)
    }
  }
}

func TestSetCookies(t *testing.T) {
  var tests = []struct {
    Id string
    cd []cookieData
  }{
    {"bill", []cookieData {
        cookieData{"type", "test"},
        cookieData{"name", "bill"},
        cookieData{"lastname", ""},
    }},
    {"bob", []cookieData {
        cookieData{"houseColor", "red"},
        cookieData{"Married", "yes"},
        cookieData{"RealPerson", "No"},
    }},
    {"", []cookieData {
        cookieData{"unicode", "ô®ôò§¹²ó­ó²"},
        cookieData{"name", "sessionless"},
    }},
  }
  for _, test := range tests {
    rc := getRedisCache()
    r := getRedisConn(rc)
    defer closeRedisConn(r)

    c := make([]*http.Cookie, 0, len(test.cd))
    for _, v := range test.cd  {
      c = append(c, &http.Cookie{Name: v.Name, Value: v.Value})
    }

    err := rc.SetCookies(test.Id, c)
    if err != nil {
      t.Errorf("Got Error: %s", err)
    }

    m, e := r.Cmd("HGETALL", getIdPath(test.Id)).Map()
    if e != nil {
      t.Errorf("Got Error: %s", e)
    }

    for _, v := range test.cd  {
      if m[v.Name] != v.Value {
        t.Errorf("Expected: %s Got: %s", v.Value, m[v.Name])
      }
    }
  }
}

func TestSetCookiesHandlesEmptyMaps(t *testing.T) {
  rc := getRedisCache()
  c := make([]*http.Cookie, 0, 0)
  err := rc.SetCookies("test", c)
  if err != nil {
    t.Errorf("Got Error: %s", err)
  }
}

func TestGetCookie(t *testing.T) {
  var tests = []struct {
      Id string
      cd cookieData
    }{
      {"bill", cookieData{"type", "test"}},
      {"bob", cookieData{"lastname", ""}},
      {"", cookieData{"name", "sessionless"}},
    }
  for _, test := range tests {
    rc := getRedisCache()
    r := getRedisConn(rc)
    defer closeRedisConn(r)

    err := r.Cmd("HSET", getIdPath(test.Id), test.cd.Name, test.cd.Value).Err
    if err != nil {
      t.Errorf("Got Error: %s", err)
      return
    }

    var c *http.Cookie
    c, err = rc.GetCookie(test.Id, test.cd.Name)
    if err != nil {
      t.Errorf("Got Error: %s", err)
      return
    }

    if c.Name != test.cd.Name {
      t.Errorf("Cookie Name: Expected: %s, Got: %s", test.cd.Name, c.Name)
    }
    if c.Value != test.cd.Value {
      t.Errorf("Cookie Value: Expected: %s, Got: %s", test.cd.Value, c.Value)
    }
  }
}

func TestGetCookies(t *testing.T) {
  var tests = []struct {
    Id string
    cd []cookieData
  }{
    {"bill", []cookieData {
        cookieData{"type", "test"},
        cookieData{"name", "bill"},
        cookieData{"lastname", ""},
    }},
    {"bob", []cookieData {
        cookieData{"houseColor", "red"},
        cookieData{"Married", "yes"},
        cookieData{"RealPerson", "No"},
    }},
    {"", []cookieData {
        cookieData{"unicode", "ô®ôò§¹²ó­ó²"},
        cookieData{"name", "sessionless"},
    }},
  }
  for _, test := range tests {
    rc := getRedisCache()
    r := getRedisConn(rc)
    defer closeRedisConn(r)

    m := map[string]string{}
    for _, v := range test.cd {
      m[v.Name] = v.Value
    }
    err := r.Cmd("HMSET", getIdPath(test.Id), m).Err
    if err != nil {
      t.Errorf("Got Error: %s", err)
      return
    }

    cookies, e := rc.GetCookies(test.Id)
    if e != nil {
      t.Errorf("Got Error: %s", err)
      return
    }

    for _, v := range cookies  {
      if m[v.Name] != v.Value {
        t.Errorf("Expected: %s Got: %s", m[v.Name], v.Value)
      }
    }
  }
}

func TestChangeCookiesId(t *testing.T) {
  var tests = []struct {
    Id1 string
    Id2 string
    cd cookieData
  }{
    {"bill", "bob", cookieData{"type", "test"}},
    {"cat", "dog", cookieData{"lastname", ""}},
    {"", "something", cookieData{"name", "sessionless"}},
  }
  for _, test := range tests {
    rc := getRedisCache()
    r := getRedisConn(rc)
    defer closeRedisConn(r)

    err := r.Cmd("HSET", getIdPath(test.Id1), test.cd.Name, test.cd.Value).Err
    if err != nil {
      t.Errorf("Got Error: %s", err)
      return
    }

    err = rc.ChangeCookiesId(test.Id1, test.Id2)
    if err != nil {
      t.Errorf("Got Error: %s", err)
      return
    }

    m, e := r.Cmd("HGETALL", getIdPath(test.Id2)).Map()
    if e != nil {
      t.Errorf("Got Error: %s", e)
      return
    }

    if len(m) < 1 {
      t.Error("No data was copied")
    }
    if m[test.cd.Name] != test.cd.Value {
      t.Errorf("Expected: %s Got: %s", m[test.cd.Name], m[test.cd.Name])
    }
  }
}
