//Copyright Dylan Enloe 2017

package kyogetsu

import (
  "errors"
  "net/http"
  "strings"
  "testing"
  )

type failReader struct {
  //doesn't need any data
}

func (r failReader) Read(b []byte) (n int, err error) {
  return 0, errors.New("Read Error")
}

func verifyRequestInfo(t *testing.T, ri RequestInfo, r RequestInfo) {
  if ri.Method != r.Method {
    t.Errorf("Method doesn't match. Expected: %s Got: %s", r.Method, ri.Method)
  }
  if ri.URL != r.URL {
    t.Errorf("URL doesn't match. Expected: %s Got: %s", r.URL, ri.URL)
  }
  if ri.Body != r.Body {
    t.Errorf("Method doesn't match. Expected: %s Got: %s", r.Body, ri.Body)
  }
}

func verifyMessage(t *testing.T, msg *Message, m Message) {
  if msg.ProductionReponse != m.ProductionReponse {
    t.Errorf("ProductionReponse doesn't match. Expected: %s Got: %s", m.ProductionReponse, msg.ProductionReponse)
  }
  if msg.StagingReponse != m.StagingReponse {
    t.Errorf("StagingReponse doesn't match. Expected: %s Got: %s", m.StagingReponse, msg.StagingReponse)
  }
  verifyRequestInfo(t, msg.Request, m.Request)
}

func TestNewRequestInfo(t *testing.T) {
  var tests = []RequestInfo {
      {"POST", "This is a test", "example.com"},
      {"GET", "this is also a test", "test.com"},
    }
  for _, test := range tests {
    b := strings.NewReader(test.Body)
    r, _ := http.NewRequest(test.Method, test.URL, b)
    ri := NewRequestInfo(r)
    verifyRequestInfo(t, ri, test)
  }
}

func TestNewRequestInfoWithoutBody(t *testing.T) {
  var tests = []RequestInfo {
      {"POST", "", "example.com"},
      {"GET", "", "test.com"},
    }
  for _, test := range tests {
    r, _ := http.NewRequest(test.Method, test.URL, failReader{})
    ri := NewRequestInfo(r)
    verifyRequestInfo(t, ri, test)
  }
}

func TestNewMessage(t *testing.T) {
  var tests = []Message {
      {RequestInfo{"POST", "This is a test", "example.com"}, "Prod", "Test"},
      {RequestInfo{"GET", "this is also a test" , "test.com"}, "", ""},
    }
  for _, test := range tests {
    b := strings.NewReader(test.Request.Body)
    r, _ := http.NewRequest(test.Request.Method, test.Request.URL, b)
    m := NewMessage(test.ProductionReponse, test.StagingReponse, r)

    verifyMessage(t, m, test)
  }
}
