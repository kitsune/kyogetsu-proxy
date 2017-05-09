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

type requestInfo struct {
  Method string
  URL string
  Body string
}

type message struct {
  Pr string
  Sr string
  r requestInfo
}

func verifyRequestInfo(t *testing.T, ri RequestInfo, r requestInfo) {
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

func verifyMessage(t *testing.T, msg *Message, m message) {
  if msg.ProductionReponse != m.Pr {
    t.Errorf("ProductionReponse doesn't match. Expected: %s Got: %s", m.Pr, msg.ProductionReponse)
  }
  if msg.StagingReponse != m.Sr {
    t.Errorf("StagingReponse doesn't match. Expected: %s Got: %s", m.Sr, msg.StagingReponse)
  }
  verifyRequestInfo(t, msg.Request, m.r)
}

func TestNewRequestInfo(t *testing.T) {
  var tests = []requestInfo {
      {"POST", "example.com", "This is a test"},
      {"GET", "test.com", "this is also a test"},
    }
  for _, test := range tests {
    b := strings.NewReader(test.Body)
    r, _ := http.NewRequest(test.Method, test.URL, b)
    ri := NewRequestInfo(r)
    verifyRequestInfo(t, ri, test)
  }
}

func TestNewRequestInfoWithoutBody(t *testing.T) {
  var tests = []requestInfo {
      {"POST", "example.com", ""},
      {"GET", "test.com", ""},
    }
  for _, test := range tests {
    r, _ := http.NewRequest(test.Method, test.URL, failReader{})
    ri := NewRequestInfo(r)
    verifyRequestInfo(t, ri, test)
  }
}

func TestNewMessage(t *testing.T) {
  var tests = []message {
      {"Prod", "Test", requestInfo{"POST", "example.com", "This is a test"}},
      {"", "", requestInfo{"GET", "test.com", "this is also a test"}},
    }
  for _, test := range tests {
    b := strings.NewReader(test.r.Body)
    r, _ := http.NewRequest(test.r.Method, test.r.URL, b)
    m := NewMessage(test.Pr, test.Sr, r)

    verifyMessage(t, m, test)
  }
}
