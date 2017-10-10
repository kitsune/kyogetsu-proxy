//Copyright Dylan Enloe 2017

package kyogetsu

import (
  "errors"
  "net/http"
  "net/http/httputil"
  "net/http/httptest"
  "net/url"
)

//ProxyHandler interface provides access to a
//production and staging ReverseProxy.
type ProxyHandler interface {
  Production(*http.Request) *httputil.ReverseProxy
  Staging(*http.Request) *httputil.ReverseProxy

}

//SingleProxyHandler is a stuct that impliments
//the ProxyHandler, providing a single production
//ReverseProxy and a single staging ReverseProxy
type SingleProxyHandler struct {
  ProductionProxy *httputil.ReverseProxy
  StagingProxy *httputil.ReverseProxy
}

//Production returns the ProductionProxy
func (p SingleProxyHandler) Production(*http.Request) *httputil.ReverseProxy {
  return p.ProductionProxy
}

//Staging returns the StagingProxy
func (p SingleProxyHandler) Staging(*http.Request) *httputil.ReverseProxy {
  return p.StagingProxy
}

//NewSingleProxyHander returns a new SingleProxyHandler by parseing
//the two URL strings given by p and s, storing them in Production
//and Staging respectively
func NewSingleProxyHandler(p string, s string) SingleProxyHandler {
  pURL, _ := url.Parse(p)
  pProxy := httputil.NewSingleHostReverseProxy(pURL)
  sURL, _ := url.Parse(s)
  sProxy := httputil.NewSingleHostReverseProxy(sURL)
  return SingleProxyHandler{ProductionProxy: pProxy, StagingProxy: sProxy}
}

//IdFunction type provides a method of determining the session
//id of the user.
type IdFunction func([]*http.Cookie) (string, error)

//CookieIdFunction generates an IdFunction.  This function will for the
//cookie whoes name matches the string provided and return it's value
//as the session id
func CookieIdFunction(s string) func([]*http.Cookie) (string, error) {
  return func(c []*http.Cookie) (string, error) {
    for i := range c {
      if c[i].Name == s {
        return c[i].Value, nil
      }
    }
    return "", errors.New("Could not find id cookie")
  }
}

//KyogetsuProxy contains all the data and functions to
//make the Kyogetsu Proxy system work.
type KyogetsuProxy struct {
  ph ProxyHandler
  ms MessageSender
  ccache CookieCache
  ignoredCookies []string
  idFunc IdFunction
}

//NewKyogetsuProxy creates a new KyogetsuProxy with the
//provided configuration
func NewKyogetsuProxy(ph ProxyHandler, ms MessageSender, c CookieCache, idf IdFunction) KyogetsuProxy {
  return KyogetsuProxy{ph: ph, ms: ms, ccache: c, idFunc: idf}
}

//ServeHTTP sends the request to the production reverse proxy
//then invokes the results to the HandleStaging method
func (p KyogetsuProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  nr, _ := http.NewRequest(r.Method, r.URL.String(), r.Body)
  nr.Header = r.Header
  pw := httptest.NewRecorder()
  p.ph.Production(r).ServeHTTP(pw, r)
  for k, v := range pw.HeaderMap {
      w.Header()[k] = v
  }
  w.WriteHeader(pw.Code)
  w.Write(pw.Body.Bytes())
  go p.HandleStaging(nr, pw)
}

//loadCookies any cookie data stored in the CookieCache
//and write it to the request, overriding any existing
//values
func (p KyogetsuProxy) loadCookies(id string, r *http.Request) error {
  sc, err := p.ccache.GetCookies(id)
  if err != nil {
    return err
  }

  r.Header.Del("Cookie")
  for _, v := range sc {
    r.AddCookie(v)
  }
  return nil
}

//saveCookies saves any cookies in the Response to the CookieCache
//if the session id is changed it will copy all the cookies
//from the old id to the new id before overwriting them
func (p KyogetsuProxy) saveCookies(id string, w http.ResponseWriter) error {
  r := http.Response{Header: w.Header()}
  c := r.Cookies()

  err := p.ccache.SetCookies(id, c)
  if err != nil {
    return err
  }
  return nil
}

//HandleStaging prepares and send the request to the staging proxy
//updates the cookies if needed and sends the results to the
//message sender
func (p KyogetsuProxy) HandleStaging(r *http.Request, pw *httptest.ResponseRecorder) {
  sr, _ := http.NewRequest(r.Method, r.URL.String(), r.Body)
  sr.Header = r.Header
  id, id_err := p.idFunc(r.Cookies())
  if id_err == nil {
    p.loadCookies(id, sr)
  }

  sw := httptest.NewRecorder()
  p.ph.Staging(r).ServeHTTP(sw, sr)

  //update id if a new id is given
  resp := http.Response{Header: pw.Header()}
  c := resp.Cookies()
  if n, e := p.idFunc(c); e == nil && n != id {
    //if the old id exists change update where the data is stored
    if id_err == nil {
      p.ccache.ChangeCookiesId(id, n)
    }
    id = n
  }

  p.saveCookies(id, sw)

  m := NewMessage(pw, sw, r, sr)
  p.ms.SendMessage(m)
}
