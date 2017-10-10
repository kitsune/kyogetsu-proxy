//Copyright Dylan Enloe 2017

package kyogetsu

import (
  "encoding/json"
  "github.com/nats-io/go-nats"
  "github.com/nats-io/gnatsd/test"
  "net/http"
  "sync/atomic"
  "testing"
  "time"
  )

func TestSendMessageBadUrl(t *testing.T) {
  ns := NewNatsSender("bad url", "subject")
  m := Message{
    RequestInfo{"POST", "bad url", http.Header{"Cookie": {"A"}}, "body"},
    RequestInfo{"POST", "bad url", http.Header{"Cookie": {"A"}}, "body"},
    ResponseInfo{200, http.Header{"Cookie": {"A"}}, "prod"},
    ResponseInfo{200, http.Header{"Cookie": {"A"}}, "test"},
  }
  err := ns.SendMessage(&m)
  if err == nil {
    t.Error("The NatsSender should return an Error for a bad URL but didn't")
  }
  if err.Error() != "nats: no servers available for connection" {
    t.Errorf("Unexpected error: %s", err)
  }
}

func TestSendMessage(t *testing.T) {
  s := test.RunDefaultServer()
  defer s.Shutdown()

  var tests = []struct {
      Subject string
      Msg Message
    }{
      {"test", Message{
        RequestInfo{"POST", "example.com", http.Header{"Header": {"Yes"}}, "testing"},
        RequestInfo{"POST", "example.com", http.Header{"Header": {"No"}}, "testing"},
        ResponseInfo{200, http.Header{}, "prod"},
        ResponseInfo{200, http.Header{}, "test"},
      }},
      {"BOB", Message{
        RequestInfo{"GET", "testing.now", http.Header{}, "what are you doing?"},
        RequestInfo{"GET", "testing.now", http.Header{}, "what are you doing?"},
        ResponseInfo{200, http.Header{"Simple": {"B"}}, "serving people"},
        ResponseInfo{200, http.Header{"Complex": {"B * 2i"}}, "testing code"},
      }},
    }
  for _, test := range tests {
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
      t.Errorf("Error connecting to NATS server: %s", err)
    }
    defer nc.Close()

    count := int32(0)
    s, err := nc.Subscribe(test.Subject, func(msg *nats.Msg) {
      atomic.AddInt32(&count, 1)
      var m Message
      json.Unmarshal(msg.Data, &m)
      verifyMessage(t, &m, test.Msg)
    })
    s.AutoUnsubscribe(1)
    defer s.Unsubscribe()

    ns := NewNatsSender("nats://localhost:4222", test.Subject)
    ns.SendMessage(&test.Msg)

    nc.Flush()
    time.Sleep(100 * time.Millisecond)

    if count < 1 {
      t.Error("Message was not recieved")
    }
  }
}
