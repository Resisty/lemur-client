package main

import (
  "fmt"
  "github.com/DataDog/datadog-go/statsd"
)

type statsdClient struct {
    prefix string
    host   string
    port   string
}

func newStatsd(s *statsdClient) *statsd.Client {
    client, err := statsd.New(fmt.Sprintf("%s:%s",
                                          s.host,
                                          s.port))
    if err != nil {
        Logs.Errorf("%+v", err)
    }
    client.Namespace = s.prefix
    return client
}

func NewCertsStatsd(host, port string) *statsd.Client {
    client := &statsdClient{"lemur.certs.", host, port}
    return newStatsd(client)
}

func NewHttpStatsd(host, port string) *statsd.Client {
    client := &statsdClient{"lemur.http.", host, port}
    return newStatsd(client)
}
