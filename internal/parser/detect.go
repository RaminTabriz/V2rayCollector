package parser

import (
    "strings"
)

// DetectProtocol نوع پروتکل کانفیگ را تشخیص می‌دهد
// پشتیبانی از: Vmess, Vless, Trojan, Shadowsocks, ShadowsocksR, Hysteria2, Tuic,
// Wireguard, WARP, MTProto, SOCKS5, SOCKS, HTTP/HTTPS, Argo, Slipnet, Invizible
func DetectProtocol(cfg string) string {
    // 1. پروتکل‌های تلگرام (اولویت بالا)
    if strings.HasPrefix(cfg, "tg://socks?") || strings.HasPrefix(cfg, "https://t.me/socks?") {
        return "telegram_socks"
    }
    if strings.HasPrefix(cfg, "tg://proxy?") || strings.Contains(cfg, "t.me/proxy?") {
        return "mtproto"
    }

    // 2. پروتکل‌های با پیشوند مشخص
    if strings.HasPrefix(cfg, "vmess://") {
        return "vmess"
    }
    if strings.HasPrefix(cfg, "vless://") {
        return "vless"
    }
    if strings.HasPrefix(cfg, "trojan://") {
        return "trojan"
    }
    if strings.HasPrefix(cfg, "ss://") {
        return "ss"
    }
    if strings.HasPrefix(cfg, "ssr://") {
        return "ssr"
    }
    if strings.HasPrefix(cfg, "hysteria2://") {
        return "hysteria2"
    }
    if strings.HasPrefix(cfg, "hy2://") {
        return "hysteria2"
    }
    if strings.HasPrefix(cfg, "tuic://") {
        return "tuic"
    }
    if strings.HasPrefix(cfg, "wireguard://") {
        return "wireguard"
    }
    if strings.HasPrefix(cfg, "warp://") {
        return "warp"
    }
    if strings.HasPrefix(cfg, "slipnet://") {
        return "slipnet"
    }

    // 3. پروکسی‌های SOCKS
    if strings.HasPrefix(cfg, "socks5://") {
        return "socks5"
    }
    if strings.HasPrefix(cfg, "socks://") {
        return "socks"
    }

    // 4. HTTP و HTTPS
    if strings.HasPrefix(cfg, "http://") {
        return "http"
    }
    if strings.HasPrefix(cfg, "https://") {
        return "https"
    }

    // 5. Argo (بلوک چندخطی)
    if strings.HasPrefix(cfg, "-----BEGIN ARGO") {
        return "argo"
    }

    // 6. Invizible Pro (obfs4 / webtunnel)
    if strings.Contains(cfg, "obfs4") || strings.Contains(cfg, "webtunnel") {
        return "invizible"
    }

    // 7. در صورت عدم تشخیص
    return "mixed"
}

// GetProtocolIcon ایموجی مربوط به هر پروتکل را برمی‌گرداند
func GetProtocolIcon(proto string) string {
    switch proto {
    case "vmess":
        return "📦"
    case "vless":
        return "🕳️"
    case "trojan":
        return "🐴"
    case "ss":
        return "🐍"
    case "ssr":
        return "🔄"
    case "hysteria2":
        return "⚡"
    case "tuic":
        return "🧩"
    case "wireguard":
        return "🔒"
    case "warp":
        return "🌌"
    case "mtproto", "telegram_socks":
        return "📱"
    case "socks5":
        return "🧦"
    case "socks":
        return "🧦"
    case "http":
        return "🌐"
    case "https":
        return "🔒"
    case "argo":
        return "☁️"
    case "slipnet":
        return "🕸️"
    case "invizible":
        return "🛡️"
    case "mixed":
        return "🎭"
    default:
        return "📄"
    }
}

// GetProtocolFileName نام فایل مناسب برای هر پروتکل (بدون پسوند) را برمی‌گرداند
func GetProtocolFileName(proto string) string {
    switch proto {
    case "vmess":
        return "📦 VMess"
    case "vless":
        return "🕳️ VLess"
    case "trojan":
        return "🐴 Trojan"
    case "ss":
        return "🐍 Shadowsocks"
    case "ssr":
        return "🔄 SSR"
    case "hysteria2":
        return "⚡ Hysteria2"
    case "tuic":
        return "🧩 Tuic"
    case "wireguard":
        return "🔒 WireGuard"
    case "warp":
        return "🌌 WARP"
    case "mtproto":
        return "📱 MTProto Proxy"
    case "telegram_socks":
        return "🧦 SOCKS5 Proxy"
    case "socks5":
        return "🧦 SOCKS5"
    case "socks":
        return "🧦 SOCKS"
    case "http":
        return "🌐 HTTP"
    case "https":
        return "🔒 HTTPS"
    case "argo":
        return "☁️ Argo"
    case "slipnet":
        return "🕸️ Slipnet"
    case "invizible":
        return "🛡️ Invizible Pro"
    case "mixed":
        return "🎭 Mixed"
    default:
        return proto
    }
}
