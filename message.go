package kyogetsu

import (
  "fmt"
  "net/http"
  "io/ioutil"
)

type MessageSender interface {
  SendMessage(*Message)
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

func NewRequestInfo(r *http.Request) RequestInfo {
  b, err := ioutil.ReadAll(r.Body);
  if  err != nil {
    return RequestInfo{Method: r.Method, URL: r.URL.String()}
  }
  return RequestInfo{Method: r.Method, Body: string(b), URL: r.URL.String()}
}

func NewMessage(p string, s string, r *http.Request) *Message {
  return &Message{Request: NewRequestInfo(r), ProductionReponse: p, StagingReponse: s}
}

type DummySender struct {
  // Doesn't need anything
}

func (d DummySender) SendMessage(m *Message) {
  fmt.Println("Returned data:")	
  fmt.Println(m.ProductionReponse)
  fmt.Println(m.StagingReponse)
}