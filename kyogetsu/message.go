//Copyright Dylan Enloe 2017

package kyogetsu

import (
  "net/http"
  "io/ioutil"
)

type MessageSender interface {
  SendMessage(*Message) error
}

type RequestInfo struct {
  Method string
  Body string
  URL string

}

type Message struct {
  Request RequestInfo
  ProductionReponse string
  StagingReponse string
}

//NewRequestInfo generates the proper RequestInfo for
//the given http.Request
func NewRequestInfo(r *http.Request) RequestInfo {
  b, err := ioutil.ReadAll(r.Body);
  if  err != nil {
    return RequestInfo{Method: r.Method, URL: r.URL.String()}
  }
  return RequestInfo{Method: r.Method, Body: string(b), URL: r.URL.String()}
}

//NewMessage creates a new message from the given data
func NewMessage(p string, s string, r *http.Request) *Message {
  return &Message{Request: NewRequestInfo(r), ProductionReponse: p, StagingReponse: s}
}
