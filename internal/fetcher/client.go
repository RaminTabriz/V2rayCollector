package fetcher

import (
    "crypto/tls"
    "net"
    "net/http"
    "net/url"
    "os"
    "time"
)

var Client *http.Client
var proxyURL string // می‌توان از flag یا env مقداردهی کرد

// Init مقداردهی اولیه HTTP client با تنظیمات timeout و proxy
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

    // پشتیبانی از proxy از محیط یا flag
    if proxyURL == "" {
        proxyURL = os.Getenv("HTTP_PROXY")
        if proxyURL == "" {
            proxyURL = os.Getenv("HTTPS_PROXY")
        }
    }
    if proxyURL != "" {
        if proxy, err := url.Parse(proxyURL); err == nil {
            transport.Proxy = http.ProxyURL(proxy)
        }
    }

    Client = &http.Client{
        Timeout:   30 * time.Second,
        Transport: transport,
    }
}

// SetProxy تنظیم دستی proxy (اختیاری)
func SetProxy(proxy string) {
    proxyURL = proxy
    // برای اعمال تغییر، باید دوباره مقداردهی کنیم
    Init()
}

// Close بستن اتصالات بیکار client
func Close() {
    if Client != nil {
        Client.CloseIdleConnections()
    }
}
