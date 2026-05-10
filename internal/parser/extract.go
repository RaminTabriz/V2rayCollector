package parser

import (
    "regexp"
    "strings"
)

// الگوی ترکیبی برای تشخیص تمام پروتکل‌های پشتیبانی شده
var combinedRegex = regexp.MustCompile(`(vmess://[A-Za-z0-9+/]+={0,2}(?:\?[^\s]*)?|vless://[^\s]+|trojan://[^@\s]+@[^\s]+|ss://[A-Za-z0-9+/]+={0,2}@[^\s]+|ssr://[A-Za-z0-9+/=]+|hysteria2://[^\s]+|hy2://[^\s]+|tuic://[^\s]+|wireguard://[^\s]+|warp://[^\s]+|slipnet://[^\s]+|tg://proxy\?[^\s]+|tg://socks\?[^\s]+|https?://[^\s]+:\d+(?:[^\s]*)?|socks(?:5)?://[^\s]+@[^\s]+|socks(?:5)?://[^\s]+:\d+|-----BEGIN ARGO VPN BRIDGE BLOCK-----[\s\S]+?-----END ARGO VPN BRIDGE BLOCK-----)`)

// ExtractAllConfigs استخراج تمام کانفیگ‌ها از یک متن
// این تابع دو مرحله انجام می‌دهد:
// 1. درج newline قبل از هر پروتکل برای جدا کردن کانفیگ‌های چسبیده
// 2. استخراج با regex و شکستن خطوط حاوی چند کانفیگ
func ExtractAllConfigs(text string) []string {
    // مرحله 1: درج newline قبل از هر پروتکل (حل مشکل چسبیدن)
    text = insertSeparators(text)

    // مرحله 2: استخراج با regex
    matches := combinedRegex.FindAllString(text, -1)

    seen := make(map[string]bool)
    var results []string

    for _, m := range matches {
        m = strings.TrimSpace(m)
        if m == "" {
            continue
        }
        // اگر هنوز newline دارد (دو کانفیگ چسبیده)، باز هم خطوط را جدا کن
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

// insertSeparators قبل از هر پروتکل یک newline اضافه می‌کند تا کانفیگ‌های چسبیده جدا شوند
func insertSeparators(text string) string {
    prefixes := []string{
        "vmess://", "vless://", "trojan://", "ss://", "ssr://",
        "hysteria2://", "hy2://", "tuic://", "wireguard://", "warp://", "slipnet://",
        "tg://proxy?", "tg://socks?",
        "socks://", "socks5://",
        "http://", "https://",
        "-----BEGIN ARGO",
    }
    for _, p := range prefixes {
        // اگر قبل از الگو newline نبود، اضافه کن
        re := regexp.MustCompile(`([^\n])` + regexp.QuoteMeta(p))
        text = re.ReplaceAllString(text, "$1\n"+p)
    }
    return text
}

// isValidConfig اعتبارسنجی ساده برای حذف لینک‌های نامرتبط (مثل لینک دانلود فایل)
func isValidConfig(cfg string) bool {
    lower := strings.ToLower(cfg)
    // حذف لینک‌های واضحاً غیرمرتبط
    if strings.Contains(lower, ".apk") || strings.Contains(lower, ".zip") ||
        strings.Contains(lower, ".rar") || strings.Contains(lower, ".7z") ||
        strings.Contains(lower, ".jpg") || strings.Contains(lower, ".jpeg") ||
        strings.Contains(lower, ".png") || strings.Contains(lower, ".gif") ||
        strings.Contains(lower, ".mp4") || strings.Contains(lower, ".webm") {
        return false
    }
    // اگر لینک HTTP ساده و بدون پارامتر خاصی است و پورت استاندارد دارد
    if (strings.HasPrefix(cfg, "http://") || strings.HasPrefix(cfg, "https://")) &&
        !strings.ContainsAny(cfg, "?@") && !strings.Contains(cfg, "/proxy") {
        return false
    }
    return true
}
