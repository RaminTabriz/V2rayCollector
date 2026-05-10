package parser

import (
    "regexp"
    "strings"
)

var combinedRegex = regexp.MustCompile(`(vmess://[A-Za-z0-9+/]+={0,2}(?:\?[^\s]*)?|vless://[^\s]+|trojan://[^@\s]+@[^\s]+|ss://[A-Za-z0-9+/]+={0,2}@[^\s]+|ssr://[A-Za-z0-9+/=]+|hysteria2://[^\s]+|hy2://[^\s]+|tuic://[^\s]+|wireguard://[^\s]+|warp://[^\s]+|slipnet://[^\s]+|tg://proxy\?[^\s]+|tg://socks\?[^\s]+|https?://[^\s]+:\d+(?:[^\s]*)?|socks(?:5)?://[^\s]+@[^\s]+|socks(?:5)?://[^\s]+:\d+|-----BEGIN ARGO VPN BRIDGE BLOCK-----[\s\S]+?-----END ARGO VPN BRIDGE BLOCK-----)`)

func ExtractAllConfigs(text string) []string {
    // Insert newlines before each protocol prefix to separate concatenated configs
    prefixes := []string{
        "vmess://", "vless://", "trojan://", "ss://", "ssr://",
        "hysteria2://", "hy2://", "tuic://", "wireguard://", "warp://", "slipnet://",
        "tg://proxy?", "tg://socks?", "socks://", "socks5://",
        "http://", "https://", "-----BEGIN ARGO",
    }
    for _, p := range prefixes {
        re := regexp.MustCompile(`([^\n])` + regexp.QuoteMeta(p))
        text = re.ReplaceAllString(text, "$1\n"+p)
    }
    matches := combinedRegex.FindAllString(text, -1)
    seen := make(map[string]bool)
    var results []string
    for _, m := range matches {
        m = strings.TrimSpace(m)
        if m == "" {
            continue
        }
        // اگر هنوز newline دارد (دو کانفیگ چسبیده)
        if strings.Contains(m, "\n") {
            for _, line := range strings.Split(m, "\n") {
                line = strings.TrimSpace(line)
                if line != "" && !seen[line] && isValidConfig(line) {
                    seen[line] = true
                    results = append(results, line)
                }
            }
        } else if !seen[m] && isValidConfig(m) {
            seen[m] = true
            results = append(results, m)
        }
    }
    return results
}

func isValidConfig(cfg string) bool {
    // حذف لینک‌های واضحاً نامرتبط
    lower := strings.ToLower(cfg)
    if strings.Contains(lower, ".apk") || strings.Contains(lower, ".zip") ||
        strings.Contains(lower, ".jpg") || strings.Contains(lower, ".png") ||
        strings.Contains(lower, ".gif") || strings.Contains(lower, ".mp4") {
        return false
    }
    // اگر لینک HTTP ساده و بدون پارامتر خاص است
    if strings.HasPrefix(cfg, "http://") || strings.HasPrefix(cfg, "https://") {
        if !strings.Contains(cfg, "?") && !strings.Contains(cfg, ":") && !strings.Contains(cfg, "/proxy") {
            return false
        }
    }
    return true
}
