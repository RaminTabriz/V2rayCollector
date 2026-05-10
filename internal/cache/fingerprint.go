package cache

import (
    "crypto/md5"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/url"
    "strings"
)

// ComputeFingerprint محاسبه اثر انگشت یکتا برای کانفیگ بر اساس پروتکل
func ComputeFingerprint(cfg, proto string) string {
    switch proto {
    case "vmess":
        return fingerprintVmess(cfg)
    case "vless", "trojan", "ss", "ssr", "hysteria2", "tuic", "wireguard", "warp", "slipnet":
        return fingerprintCredentialURL(cfg)
    default:
        hash := md5.Sum([]byte(cfg))
        return fmt.Sprintf("%x", hash)
    }
}

// fingerprintVmess استخراج فیلدهای کلیدی از JSON داخل vmess://
func fingerprintVmess(vmessUrl string) string {
    parts := strings.SplitN(vmessUrl, "vmess://", 2)
    if len(parts) != 2 {
        return ""
    }
    decoded, err := base64.StdEncoding.DecodeString(parts[1])
    if err != nil {
        return ""
    }
    var data map[string]interface{}
    if err := json.Unmarshal(decoded, &data); err != nil {
        return ""
    }
    // ترکیب فیلدهایی که هویت سرور را مشخص می‌کنند
    add := fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s:%s",
        getStr(data, "add"), getStr(data, "port"), getStr(data, "id"),
        getStr(data, "net"), getStr(data, "type"), getStr(data, "host"),
        getStr(data, "path"), getStr(data, "tls"), getStr(data, "sni"))
    hash := md5.Sum([]byte(add))
    return fmt.Sprintf("%x", hash)
}

// fingerprintCredentialURL برای پروتکل‌های با فرمت URL استاندارد (user:pass@host:port)
func fingerprintCredentialURL(cfg string) string {
    u, err := url.Parse(cfg)
    if err != nil {
        hash := md5.Sum([]byte(cfg))
        return fmt.Sprintf("%x", hash)
    }
    userPass := ""
    if u.User != nil {
        userPass = u.User.String()
    }
    host := u.Hostname()
    port := u.Port()
    hash := md5.Sum([]byte(host + ":" + port + ":" + userPass))
    return fmt.Sprintf("%x", hash)
}

// getStr کمک‌کننده برای استخراج string از map
func getStr(data map[string]interface{}, key string) string {
    if val, ok := data[key]; ok && val != nil {
        return fmt.Sprintf("%v", val)
    }
    return ""
}
