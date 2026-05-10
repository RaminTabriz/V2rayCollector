package parser

import (
    "net/url"
    "strings"
)

func IsSecure(cfg, proto string) bool {
    switch proto {
    case "vmess":
        return isVmessSecure(cfg)
    case "vless":
        return isVlessSecure(cfg)
    case "hysteria2":
        return isHysteria2Secure(cfg)
    default:
        return true
    }
}

func isVmessSecure(vmessUrl string) bool {
    parts := strings.SplitN(vmessUrl, "vmess://", 2)
    if len(parts) != 2 {
        return true
    }
    decoded, err := base64.StdEncoding.DecodeString(parts[1])
    if err != nil {
        return true
    }
    var data map[string]interface{}
    if err := json.Unmarshal(decoded, &data); err != nil {
        return true
    }
    tlsVal, _ := data["tls"].(string)
    if tlsVal == "" {
        return true
    }
    if tlsVal != "tls" && tlsVal != "xtls" {
        return false
    }
    insecure, _ := data["allowInsecure"].(bool)
    return !insecure
}

func isVlessSecure(vlessUrl string) bool {
    u, err := url.Parse(vlessUrl)
    if err != nil {
        return true
    }
    security := u.Query().Get("security")
    allowInsecure := u.Query().Get("allowInsecure")
    encryption := u.Query().Get("encryption")
    
    if strings.ToLower(encryption) == "none" {
        return false
    }
    secureProtocols := map[string]bool{"tls": true, "reality": true, "xtls": true}
    if !secureProtocols[security] {
        return false
    }
    return allowInsecure != "1" && allowInsecure != "true"
}

func isHysteria2Secure(hy2Url string) bool {
    u, err := url.Parse(hy2Url)
    if err != nil {
        return true
    }
    insecure := u.Query().Get("insecure")
    allowInsecure := u.Query().Get("allowInsecure")
    if insecure == "1" || insecure == "true" || allowInsecure == "1" || allowInsecure == "true" {
        return false
    }
    return true
}
