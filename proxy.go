package kyogetsu

import (
  "net/http"
  "net/http/httputil"
  "net/url"
  "sync"
)

type ProxyHandler interface {
  Production(*http.Request) *httputil.ReverseProxy
  Staging(*http.Request) *httputil.ReverseProxy

}

type SingleProxyHandler struct {
  ProductionProxy *httputil.ReverseProxy
  StagingProxy *httputil.ReverseProxy
}

func (p SingleProxyHandler) Production(*http.Request) *httputil.ReverseProxy {
  return p.ProductionProxy
}

func (p SingleProxyHandler) Staging(*http.Request) *httputil.ReverseProxy {
  return p.StagingProxy
}

func NewSingleProxyHander(p string, s string) SingleProxyHandler {
  pURL, _ := url.Parse(p)
  pProxy := httputil.NewSingleHostReverseProxy(pURL)
  sURL, _ := url.Parse(s)
  sProxy := httputil.NewSingleHostReverseProxy(sURL)
  return SingleProxyHandler{ProductionProxy: pProxy, StagingProxy: sProxy}

}

type KyogetsuConfig struct {
  ProxyHandler ProxyHandler
  MessageSender MessageSender
}

type KyogetsuProxy struct {
  config KyogetsuProxy
  cm sync.RWMutex
}

func (p KyogetsuProxy) Configure(c KyogetsuConfig) {
  p.cm.Lock()
  defer p.cm.Unlock()
  p.config = c
}

func (p KyogetsuProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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