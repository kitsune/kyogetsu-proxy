//Copyright Dylan Enloe 2017

package kyogetsu

import (
  "errors"
  "net/http"
  "net/http/httptest"
  "strings"
  "testing"
  "fmt"
  )

type failReader struct {
  //doesn't need any data
}

func (r failReader) Read(b []byte) (n int, err error) {
  return 0, errors.New("Read Error")
}

func verifyHeader(t *testing.T, hi http.Header, h http.Header) {
  if len(hi) != len(h) {
    t.Errorf("Header lengths do not match. Expected: %s Got: %s", len(h), len(hi))
    fmt.Println(hi)
    return
  }
  for k, v := range hi {
    if len(v) != len(h[k]) {
      t.Errorf("Value lengths for header key %s do not match. Expected: %s Got: %s",
               k,
               len(h[k]),
               len(v))
      return
    }
    for i, val := range v {
      if val != h[k][i] {
        t.Errorf("Header field doesn't match. Expected: %s Got: %s", h[k][i], val)
      }
    }
  }
}

func verifyRequestInfo(t *testing.T, ri RequestInfo, r RequestInfo) {
  if ri.Method != r.Method {
    t.Errorf("Method doesn't match. Expected: %s Got: %s", r.Method, ri.Method)
  }
  if ri.URI != r.URI {
    t.Errorf("URI doesn't match. Expected: %s Got: %s", r.URI, ri.URI)
  }
  if ri.Body != r.Body {
    t.Errorf("Body doesn't match. Expected: %s Got: %s", r.Body, ri.Body)
  }
  verifyHeader(t, ri.Header, r.Header)
}

func verifyResponseInfo(t *testing.T, ri ResponseInfo, r ResponseInfo) {
  if ri.Status != r.Status {
    t.Errorf("Respose Status doesn't match. Expected: %s Got: %s", r.Status, ri.Status)
  }
  if ri.Body != r.Body {
    t.Errorf("Respose Body doesn't match. Expected: %s Got: %s", r.Body, ri.Body)
  }
}

func verifyMessage(t *testing.T, msg *Message, m Message) {
  verifyResponseInfo(t, msg.ProdReponse, m.ProdReponse)
  verifyResponseInfo(t, msg.StagingReponse, m.StagingReponse)
  verifyRequestInfo(t, msg.ProdRequest, m.ProdRequest)
  verifyRequestInfo(t, msg.StagingRequest, m.StagingRequest)
}

func makeResponse(t *testing.T, r *ResponseInfo) *httptest.ResponseRecorder{
  res := httptest.NewRecorder()
  res.HeaderMap = r.Header
  res.WriteString(r.Body)
  res.Code = r.Status
  return res
}

func TestNewRequestInfo(t *testing.T) {
  var tests = []RequestInfo {
      {"POST", "example.com", http.Header{"Cookie": {"A Cookie"}}, "This is a test"},
      {"GET", "test.com", http.Header{"Cookie": {"Session"}}, "this is also a test"},
    }
  for _, test := range tests {
    b := strings.NewReader(test.Body)
    r, _ := http.NewRequest(test.Method, test.URI, b)
    r.Header = test.Header
    ri := NewRequestInfo(r)
    verifyRequestInfo(t, ri, test)
  }
}

func TestNewRequestInfoWithoutBody(t *testing.T) {
  var tests = []RequestInfo {
      {"POST", "example.com", http.Header{"Cookie": {""}}, ""},
      {"GET", "test.com", http.Header{"Cookie": {"cat", "dog"}}, ""},
    }
  for _, test := range tests {
    r, _ := http.NewRequest(test.Method, test.URI, failReader{})
    r.Header = test.Header
    ri := NewRequestInfo(r)
    verifyRequestInfo(t, ri, test)
  }
}

func TestNewMessage(t *testing.T) {
  var tests = []Message {
      {
        RequestInfo{"POST", "example.com", http.Header{"Cookie": {"Prod Request"}}, "This is a prod test"},
        RequestInfo{"POST", "example.com", http.Header{"Cookie": {"Staging Resquest"}}, "This is a staging test"},
        ResponseInfo{301, http.Header{"Cookie": {"Prod Response"}}, "Prod"},
        ResponseInfo{302, http.Header{"Cookie": {"Staging Response"}}, "Test"},
      },
      {
        RequestInfo{"GET", "test.com", http.Header{}, "this is also a prod test" },
        RequestInfo{"GET", "test.com", http.Header{"Size": {"Not Empty"}}, "this is also a staging test" },
        ResponseInfo{200, http.Header{"Cookie": {"A"}}, ""},
        ResponseInfo{404, http.Header{"Cookie": {"A"}}, ""},
      },
    }
  for _, test := range tests {
    b := strings.NewReader(test.ProdRequest.Body)
    pr, _ := http.NewRequest(test.ProdRequest.Method, test.ProdRequest.URI, b)
    b = strings.NewReader(test.StagingRequest.Body)
    pr.Header = test.ProdRequest.Header
    sr, _ := http.NewRequest(test.StagingRequest.Method, test.StagingRequest.URI, b)
    sr.Header = test.StagingRequest.Header
    p := makeResponse(t, &test.ProdReponse)
    s := makeResponse(t, &test.StagingReponse)
    m := NewMessage(p, s, pr, sr)

    verifyMessage(t, m, test)
  }
}
