package fetcher

import (
    "crypto/tls"
    "net"
    "net/http"
    "time"
)

var Client *http.Client

func Init() {
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
        DialContext: (&net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
        MaxIdleConns:        100,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
    }
    // Proxy support from env
    Client = &http.Client{
        Timeout:   30 * time.Second,
        Transport: transport,
    }
}

func Close() {
    if Client != nil {
        Client.CloseIdleConnections()
    }
}
