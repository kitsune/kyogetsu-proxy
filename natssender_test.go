//Copyright Dylan Enloe 2017

package kyogetsu_test

import (
  "encoding/json"
  "github.com/nats-io/go-nats"
  "github.com/nats-io/gnatsd/test"
  "kyogetsu"
  "strings"
  "sync/atomic"
  "testing"
  "time"
  "net/http"
  )

func MakeTestMessage(m message) *kyogetsu.Message{
  b := strings.NewReader(m.r.Body)
  r, _ := http.NewRequest(m.r.Method, m.r.URL, b)
  return kyogetsu.NewMessage(m.Pr, m.Sr, r)
}

func TestSendMessageBadUrl(t *testing.T) {
  ns := kyogetsu.NewNatsSender("bad url", "subject")
  m := MakeTestMessage(message{"prod", "test", requestInfo{"POST", "bad url", "body"}})
  err := ns.SendMessage(m)
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
      Msg message
    }{
      {"test", message{
        "prod",
        "test",
        requestInfo{"POST", "example.com", "testing"},
      }},
      {"BOB", message{
        "serving people",
        "testing code",
        requestInfo{"GET", "testing.now", "what are you doing?"},
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
      var m kyogetsu.Message
      json.Unmarshal(msg.Data, &m)
      VerifyMessage(t, &m, test.Msg)
    })
    s.AutoUnsubscribe(1)
    defer s.Unsubscribe()

    ns := kyogetsu.NewNatsSender("nats://localhost:4222", test.Subject)
    m := MakeTestMessage(test.Msg)
    ns.SendMessage(m)

    nc.Flush()
    time.Sleep(100 * time.Millisecond)

    if count < 1 {
      t.Error("Message was not recieved")
    }
  }
}
