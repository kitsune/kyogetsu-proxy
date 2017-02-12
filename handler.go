package kyogetsu

import (
  "net/http"
  "net/http/httputil"
  "net/http/httptest"
  "time"
  "fmt"
)

func HandleStaging(ms MessageSender, s *httputil.ReverseProxy, r *http.Request, ps string) {
  time.Sleep(15 * time.Second)
  sw := httptest.NewRecorder()
  s.ServeHTTP(sw, r)

  sr := http.Response{Header: sw.Header()}
  fmt.Println("Response Header")
  //fmt.Println(sr.Cookies())
  for _, e := range sr.Cookies() {
      fmt.Println(e.Name)
  }
  c := http.Cookie{Name: "testing", Value: "injection"}
  r.Header.Set("Cookie", c.String())
  fmt.Println("Request Header")
  fmt.Println(r.Cookies())
  m := NewMessage(ps, sw.Body.String(), r)
  ms.SendMessage(m)
}

func NewHandleFunc(p ProxyHandler, ms MessageSender) func(w http.ResponseWriter, r *http.Request){
  return func(w http.ResponseWriter, r *http.Request) {
    //Copy the request into a new Request so it doesn't go away when
    //this function returns
    nr, _ := http.NewRequest(r.Method, r.URL.String(), r.Body)
    nr.Header = r.Header 
    
    pw := httptest.NewRecorder()
    p.Production(r).ServeHTTP(pw, r)
    for k, v := range pw.HeaderMap {
        w.Header()[k] = v
    }
    w.WriteHeader(pw.Code)
    pw.Body.WriteTo(w)

    go HandleStaging(ms, p.Staging(r), nr, pw.Body.String())
  }
}