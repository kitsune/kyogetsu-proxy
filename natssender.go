package kyogetsu

import (
  "encoding/json"
  "github.com/nats-io/go-nats"
  "log"
)

type NatsSender struct {
  URLStr string
  PubSubj string
}

func (n NatsSender) SendMessage(m *Message) {
  nc, err := nats.Connect(n.URLStr)
  if err != nil {
    log.Println(err)
    return
  }
  defer nc.Close()
  b, err := json.Marshal(m)
  if err != nil {
    log.Println("Failed to encode the following message")
    log.Println(err)
    return
  }
  
  nc.Publish(n.PubSubj, b)
  nc.Flush()

  if err := nc.LastError(); err != nil {
    log.Println("Failed to send the message: %s", b)
    log.Println(err)
  }
}

func NewNatsSender(url string, subject string) NatsSender {
  return NatsSender{URLStr: url, PubSubj: subject}
}