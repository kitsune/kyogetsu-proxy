/* Copyright Dylan Enloe 2017
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package kyogetsu

import (
  "encoding/json"
  "github.com/nats-io/go-nats"
  "log"
)

//NatsSender is an implementation of the MessageSender
//interface using go-nats
type NatsSender struct {
  URLStr string
  PubSubj string
}

//SendMessage publishes the message to the go-nats server
func (n NatsSender) SendMessage(m *Message) error{
  nc, err := nats.Connect(n.URLStr)
  if err != nil {
    log.Println(err)
    return err
  }
  defer nc.Close()
  b, err := json.Marshal(m)
  if err != nil {
    log.Println("Failed to encode the following message")
    log.Println(err)
    return err
  }

  nc.Publish(n.PubSubj, b)
  nc.Flush()

  if err := nc.LastError(); err != nil {
    log.Println("Failed to send the message: %s", b)
    log.Println(err)
    return err
  }
  return nil
}


//NewNatsSender creates a new NatsSender for the publishing
//subject queue provided
func NewNatsSender(url string, subject string) NatsSender {
  return NatsSender{URLStr: url, PubSubj: subject}
}
