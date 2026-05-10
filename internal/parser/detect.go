package parser

import (
    "strings"
)

func DetectProtocol(cfg string) string {
    // Telegram specific
    if strings.HasPrefix(cfg, "tg://socks?") || strings.HasPrefix(cfg, "https://t.me/socks?") {
        return "telegram_socks"
    }
    if strings.HasPrefix(cfg, "tg://proxy?") || strings.Contains(cfg, "t.me/proxy?") {
        return "mtproto"
    }
    // Protocol prefixes
    if strings.HasPrefix(cfg, "hy2://") {
        return "hysteria2"
    }
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
    if strings.HasPrefix(cfg, "socks://") || strings.HasPrefix(cfg, "socks5://") {
        return "socks"
    }
    if strings.HasPrefix(cfg, "http://") || strings.HasPrefix(cfg, "https://") {
        return "http"
    }
    if strings.HasPrefix(cfg, "-----BEGIN ARGO") {
        return "argo"
    }
    if strings.Contains(cfg, "obfs4") || strings.Contains(cfg, "webtunnel") {
        return "invizible"
    }
    return "mixed"
}
