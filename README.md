# Kyogetsu Proxy
Kyogetsu proxy is a lightweight reverse proxy that forks traffic between Production and Staging servers and reporting the results.  This allows a RC to experience dynamic production traffic without exposing your users to staging code.

## Features:
* Independent Production and Staging requests.  Users never have to wait on staging to finish
* Saving of Staging's cookies for subsequent requests
* Redis integration for the persistant storage of cookies.
* Publishing of results to a message queue so other programs can looks for difference (this is not done in the proxy to keep it lightweight)
* NATS integration for the message queue.
* Heavy use of interfaces so people that don't want to use NATS or Redis can implement their own prefered choice

## Installation
    go get github.com/kitsune/kyogestu-proxy

## Quick Start Example

```go
import (
  "github.com/kitsune/kyogestu-proxy/kyogetsu"
  "net/http"
)

func main() {
    ph := kyogetsu.NewSingleProxyHandler("http://127.0.0.1:8082", "http://127.0.0.1:8081")
    c := kyogetsu.NewRedisCache("127.0.0.1:6379")
    ms := kyogetsu.NewNatsSender("nats://localhost:4222", "test")
    p := kyogetsu.NewKyogetsuProxy(ph, ms, c, kyogetsu.CookieIdFunction("id"))
    h := http.NewServeMux()
    h.Handle("/", p)
    http.ListenAndServe(":8080", h)
}

```

## Libraries Used

* Redis: [Radix.v2](https://github.com/mediocregopher/radix.v2)
* NATS: [Go-NATS](https://github.com/nats-io/go-nats)
