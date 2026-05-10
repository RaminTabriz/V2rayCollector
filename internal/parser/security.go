package parser

import (
    "encoding/base64"
    "encoding/json"
    "net/url"
    "strings"
)

// IsSecure بررسی می‌کند آیا کانفیگ امن است یا نه
// کانفیگ‌های ناامن (allowInsecure, encryption=none, insecure=1) حذف می‌شوند
func IsSecure(cfg, proto string) bool {
    switch proto {
    case "vmess":
        return isVmessSecure(cfg)
    case "vless":
        return isVlessSecure(cfg)
    case "trojan":
        return isTrojanSecure(cfg)
    case "ss", "ssr":
        return isShadowsocksSecure(cfg)
    case "hysteria2":
        return isHysteria2Secure(cfg)
    case "tuic":
        return isTuicSecure(cfg)
    case "wireguard":
        return isWireguardSecure(cfg)
    case "mtproto", "telegram_socks", "socks", "socks5", "http", "https", "argo", "slipnet", "invizible", "mixed", "warp":
        // این پروتکل‌ها یا فاقد پارامتر امنیتی هستند یا امنیت آن‌ها خارج از این ماژول بررسی می‌شود
        return true
    default:
        return true
    }
}

// isVmessSecure بررسی امنیت کانفیگ VMess
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
    // بررسی tls
    tlsVal, _ := data["tls"].(string)
    if tlsVal == "" {
        // اگر tls مشخص نباشد، معمولاً امن نیست
        return false
    }
    if tlsVal != "tls" && tlsVal != "xtls" {
        return false
    }
    // بررسی allowInsecure
    insecure, _ := data["allowInsecure"].(bool)
    if insecure {
        return false
    }
    return true
}

// isVlessSecure بررسی امنیت کانفیگ VLess
func isVlessSecure(vlessUrl string) bool {
    u, err := url.Parse(vlessUrl)
    if err != nil {
        return true
    }
    security := u.Query().Get("security")
    allowInsecure := u.Query().Get("allowInsecure")
    encryption := u.Query().Get("encryption")

    // encryption=none به معنی بدون رمزنگاری است
    if strings.ToLower(encryption) == "none" {
        return false
    }
    // پروتکل‌های امن مجاز
    secureProtocols := map[string]bool{"tls": true, "reality": true, "xtls": true}
    if !secureProtocols[security] {
        return false
    }
    // allowInsecure=1 یا true ممنوع
    if allowInsecure == "1" || allowInsecure == "true" {
        return false
    }
    return true
}

// isTrojanSecure بررسی امنیت کانفیگ Trojan
func isTrojanSecure(trojanUrl string) bool {
    u, err := url.Parse(trojanUrl)
    if err != nil {
        return true
    }
    security := u.Query().Get("security")
    allowInsecure := u.Query().Get("allowInsecure")
    // Trojan معمولاً به TLS نیاز دارد
    if security != "" && security != "tls" && security != "xtls" {
        return false
    }
    if allowInsecure == "1" || allowInsecure == "true" {
        return false
    }
    return true
}

// isShadowsocksSecure بررسی امنیت Shadowsocks/SSR (معمولاً امن هستند مگر پارامتر خاصی داشته باشند)
func isShadowsocksSecure(ssUrl string) bool {
    // Shadowsocks ذاتاً رمزنگاری دارد،但 برخی پارامترها می‌توانند ناامن کنند
    // بررسی ساده: اگر cipher ضعیف باشد (اختیاری)
    if strings.Contains(ssUrl, "table") || strings.Contains(ssUrl, "rc4") {
        return false
    }
    return true
}

// isHysteria2Secure بررسی امنیت Hysteria2
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

// isTuicSecure بررسی امنیت Tuic
func isTuicSecure(tuicUrl string) bool {
    u, err := url.Parse(tuicUrl)
    if err != nil {
        return true
    }
    allowInsecure := u.Query().Get("allow_insecure")
    if allowInsecure == "1" || allowInsecure == "true" {
        return false
    }
    return true
}

// isWireguardSecure بررسی امنیت WireGuard (معمولاً امن است)
func isWireguardSecure(wgUrl string) bool {
    // WireGuard پیش‌فرض امن است، پارامتر خاصی برای غیرامن کردن ندارد
    return true
}
