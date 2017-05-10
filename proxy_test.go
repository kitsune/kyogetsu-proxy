//Copyright Dylan Enloe 2017

package kyogetsu

import (
  "fmt"
  "net/http"
  "net/http/httptest"
  "strings"
  "testing"
  )

//A dummy MessageSender that checks the messages passed to it
type dummySender struct {
  expected Message
  t *testing.T
}

func (d dummySender) SendMessage(m *Message) error{
  verifyMessage(d.t, m, d.expected)
  return nil
}

func newTestRequest() *http.Request {
  r, _ := http.NewRequest("POST", "", strings.NewReader("this is a test"))
  return r
}

func newTestServer(s string) *httptest.Server {
  handler := func( w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, s)
  }
  return httptest.NewServer(http.HandlerFunc(handler))
}

//Return a server that sets some cookies
func newCookieServer(s string, c []*http.Cookie) *httptest.Server {
  handler := func( w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, s)
    for _, cookie := range c {
      http.SetCookie(w, cookie)
    }
  }
  return httptest.NewServer(http.HandlerFunc(handler))
}

func newStagingServer(c...*http.Cookie) *httptest.Server {
  return newCookieServer("Staging", c)
}

func newProdServer(c...*http.Cookie) *httptest.Server {
  return newCookieServer("Prod", c)
}

func newTestKyogetsuProxy(p *httptest.Server, s *httptest.Server, ms MessageSender, c *RedisCache) KyogetsuProxy {
  ph := NewSingleProxyHandler(p.URL, s.URL)
  return NewKyogetsuProxy(ph, ms, *c, CookieIdFunction("id"))
}

func setBasicCookie(r *http.Request) {
  r.AddCookie(&http.Cookie{Name: "key", Value: "bob"})
}

func TestNewSingleProxyHandler(t *testing.T) {
  tests := []struct {
    prodURL string
    stageURL string
  } {
    {"http://example.com/", "http://test.example.com/"},
    {"https://place1.com:50/", "ftp://place2.com:50/"},
    {"prod://test.org:8081/", "test://test.org:8082/"},
  }
  for _, test := range tests {
    sp := NewSingleProxyHandler(test.prodURL, test.stageURL)

    r := newTestRequest()

    sp.ProductionProxy.Director(r)
    if r.URL.String() != test.prodURL {
      t.Error(r.URL.Path)
      t.Errorf("Production URLs don't match. Expecting: %s Got: %s", test.prodURL, r.URL.String())
    }

    sp.StagingProxy.Director(r)
    if r.URL.String() != test.stageURL {
      t.Error(r.URL.Path)
      t.Errorf("Production URLs don't match. Expecting: %s Got: %s", test.stageURL, r.URL.String())
    }
  }
}

func TestSingleProxyProduction(t *testing.T) {
  sp := NewSingleProxyHandler("https://testing.prod.com", "https://testing.staging.com")
  r := newTestRequest()
  if sp.Production(r) != sp.ProductionProxy {
    t.Errorf("Expecting: %s Got: %s", sp.Production(r), sp.ProductionProxy )
  }
}

func TestSingleProxyStaging(t *testing.T) {
  sp := NewSingleProxyHandler("https://testing.prod.com", "https://testing.staging.com")
  r := newTestRequest()
  if sp.Staging(r) != sp.StagingProxy {
    t.Errorf("Expecting: %s Got: %s", sp.Staging(r), sp.StagingProxy )
  }
}

func TestSingleProxyProductionIsNotStaging(t *testing.T) {
  sp := NewSingleProxyHandler("https://testing.prod.com", "https://testing.prod.com")
  if sp.ProductionProxy == sp.StagingProxy {
    t.Error("Staging and Production should be different reverse proxies")
  }
}

func TestCookieIdFunction(t *testing.T) {
  tests := []struct {
    Target int
    Cookies []*http.Cookie
  }{
    {0, []*http.Cookie{
      &http.Cookie{Name: "key", Value: "bob"},
      &http.Cookie{Name: "ke", Value: "bill"},
    }},
    {1, []*http.Cookie{
      &http.Cookie{Name: "type", Value: "test"},
      &http.Cookie{Name: "id", Value: "betty"},
    }},
    {1, []*http.Cookie{
      &http.Cookie{Name: "prod", Value: "true"},
      &http.Cookie{Name: "id", Value: "12345"},
      &http.Cookie{Name: "secret", Value: "there isn't one"},
    }},
  }
  for _, test := range tests {
    f := CookieIdFunction(test.Cookies[test.Target].Name)
    id, err := f(test.Cookies)
    if err != nil {
      t.Errorf("Unexpected Error: %s", err)
    }
    if id != test.Cookies[test.Target].Value {
      t.Errorf("Expected %s Got: %s", test.Cookies[test.Target].Value, id)
    }
  }
}

func TestCookieIdFunctionNoKey(t *testing.T) {
  c := []*http.Cookie {
    &http.Cookie{Name: "key", Value: "bob"},
    &http.Cookie{Name: "ke", Value: "bill"},
    }
  f := CookieIdFunction("id")
  _, err := f(c)
  if err == nil {
    t.Error("Key not Found Error was not returned")
  }
}

func TestLoadCookies(t *testing.T) {
  tests := []struct{
    Id string
    Name string
    Value string
  } {
    {"bill", "id", "12345"},
    {"bob", "cookie", ""},
  }
  for _, test := range tests {
    r := newTestRequest()
    setBasicCookie(r)

    ps := newProdServer()
    defer ps.Close()

    ss := newStagingServer()
    defer ss.Close()

    rc := getRedisCache()
    k := newTestKyogetsuProxy(ps, ss, dummySender{}, rc)
    rc.pool.Cmd("FLUSHALL")
    err := k.ccache.SetCookie(test.Id, &http.Cookie{Name: test.Name, Value: test.Value})
    if err != nil {
      t.Errorf("Got Error: %s", err)
    }

    k.loadCookies(test.Id, r)
    c := r.Cookies()
    if c[0].Name != test.Name {
      t.Errorf("Cookie Name Mismatch Expected: %s Got: %s", test.Name, c[0].Name)
    }
    if c[0].Value != test.Value {
      t.Errorf("Cookie Value Mismatch Expected: %s Got: %s", test.Value, c[0].Value)
    }
  }
}

func TestSaveCookies(t *testing.T) {
  tests := []struct{
    Id string
    Name string
    Value string
  } {
    {"id", "method", "POST"},
    {"sessionid", "type", ""},
  }
  for _, test := range tests {
    r := httptest.NewRecorder()
    http.SetCookie(r, &http.Cookie{Name: test.Name, Value: test.Value})

    ps := newProdServer()
    defer ps.Close()

    ss := newStagingServer()
    defer ss.Close()

    rc := getRedisCache()
    k := newTestKyogetsuProxy(ps, ss, dummySender{}, rc)
    rc.pool.Cmd("FLUSHALL")

    k.saveCookies(test.Id, r)
    c, err := rc.GetCookies(test.Id)
    if err != nil {
      t.Errorf("Unexpected Error: %s", err)
    }
    if c[0].Name != test.Name {
      t.Errorf("Cookie Name Mismatch Expected: %s Got: %s", test.Name, c[0].Name)
    }
    if c[0].Value != test.Value {
      t.Errorf("Cookie Value Mismatch Expected: %s Got: %s", test.Value, c[0].Value)
    }
  }
}

func TestHandleStaging(t *testing.T) {
  tests := []struct{
    IdKey string
    ProdC []*http.Cookie
    StagingC []*http.Cookie
    NewProdC []*http.Cookie
    NewStagingC []*http.Cookie
  } {
    {"id", nil, nil, nil, nil},
    {"id", []*http.Cookie{
      &http.Cookie{Name: "id", Value: "bob"},
      }, nil, nil, nil,},
    {"id", []*http.Cookie{
        &http.Cookie{Name: "id", Value: "bob"},
        &http.Cookie{Name: "server", Value: "prod"},
      }, []*http.Cookie{
        &http.Cookie{Name: "server", Value: "test"},
      }, nil, nil,},
    {"id", []*http.Cookie{
        &http.Cookie{Name: "id", Value: "bob"},
      }, nil, []*http.Cookie{
        &http.Cookie{Name: "id", Value: "bill"},
      }, nil,},
    {"id", []*http.Cookie{
        &http.Cookie{Name: "id", Value: "bob"},
        &http.Cookie{Name: "server", Value: "prod"},
      }, []*http.Cookie{
        &http.Cookie{Name: "server", Value: "test"},
      }, []*http.Cookie{
        &http.Cookie{Name: "new", Value: "value"},
      }, []*http.Cookie{
        &http.Cookie{Name: "new", Value: "cookie"},
      },},
  }
  for _, test := range tests {
    ps := newProdServer(test.ProdC...)
    defer ps.Close()

    ss := newStagingServer(test.StagingC...)
    defer ss.Close()

    r, _ := http.NewRequest("Post", ps.URL + "/", strings.NewReader(""))
    for _, c := range test.ProdC {
      r.AddCookie(c)
    }

    nr, _ := http.NewRequest(r.Method, ss.URL + "/", r.Body)
    for _, c := range test.StagingC {
      nr.AddCookie(c)
    }

    ms := dummySender{Message{NewRequestInfo(nr), "Prod", "Staging"}, t}
    rc := getRedisCache()
    k := newTestKyogetsuProxy(ps, ss, ms, rc)

    pw := httptest.NewRecorder()
    k.ph.Production(r).ServeHTTP(pw, r)

    k.HandleStaging(nr, pw)
  }
}

func TestServeHTTP(t *testing.T) {
  ps := newProdServer()
  defer ps.Close()

  ss := newStagingServer()
  defer ss.Close()

  r, _ := http.NewRequest("Post", ps.URL + "/", strings.NewReader(""))
  nr, _ := http.NewRequest(r.Method, ss.URL + "/", r.Body)

  ms := dummySender{Message{NewRequestInfo(nr), "Prod", "Staging"}, t}
  rc := getRedisCache()
  k := newTestKyogetsuProxy(ps, ss, ms, rc)

  pw := httptest.NewRecorder()
  k.ph.Production(r).ServeHTTP(pw, r)

  if pw.Body.String() != "Prod" {
    t.Errorf("Expected: Prod Got: %s", pw.Body.String())
  }
}
