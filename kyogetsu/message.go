//Copyright Dylan Enloe 2017

package kyogetsu

import (
  "net/http"
  "net/http/httptest"
  "io/ioutil"
)

type MessageSender interface {
  SendMessage(*Message) error
}

type RequestInfo struct {
  Method string
  URI string
  Header http.Header
  Body string
}

type ResponseInfo struct {
  Status int
  Header http.Header
  Body string
}

type Message struct {
  ProdRequest RequestInfo
  StagingRequest RequestInfo
  ProdReponse ResponseInfo
  StagingReponse ResponseInfo
}

//NewRequestInfo generates the proper RequestInfo for
//the given http.Request
func NewRequestInfo(r *http.Request) RequestInfo {
  b, err := ioutil.ReadAll(r.Body)
  if  err != nil {
    return RequestInfo{Method: r.Method, URI: r.URL.String(), Header: r.Header}
  }
  return RequestInfo{Method: r.Method, URI: r.URL.String(), Header: r.Header, Body: string(b)}
}

//NewResponseInfo generates the proper ResponseInfo for
//the given httptest.ResponseRecorder
func NewResponseInfo(r *httptest.ResponseRecorder) ResponseInfo {
  return ResponseInfo{Status: r.Code, Header: r.Header(), Body: r.Body.String()}
}

//NewMessage creates a new message from the given data
func NewMessage(p *httptest.ResponseRecorder, s *httptest.ResponseRecorder,
      pr *http.Request, sr *http.Request) *Message {
  return &Message{
    ProdRequest: NewRequestInfo(pr),
    StagingRequest: NewRequestInfo(sr),
    ProdReponse: NewResponseInfo(p),
    StagingReponse: NewResponseInfo(s)}
}
