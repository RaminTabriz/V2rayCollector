package output

import (
    "os"
    "path/filepath"
    "sort"
    "strings"

    "github.com/RaminTabriz/V2rayCollector/internal/cache"
    "github.com/RaminTabriz/V2rayCollector/internal/parser"
    "github.com/projectdiscovery/gologger"
)

// WriteAllConfigs نوشتن کانفیگ‌های جدید (از آخرین آرشیو به بعد) در پوشه 📦 all_configs
// فایل‌ها در زیرپوشه‌های 📡 telegram و 🔗 subscription و بر اساس پروتکل ذخیره می‌شوند
func WriteAllConfigs(c *cache.Cache, sortOutput bool) {
    entries := c.GetAll()
    lastArch := GetLastArchiveTime()

    // ساختار: منبع -> نام فایل -> لیست کانفیگ
    sourceFiles := map[string]map[string][]string{
        "telegram":     {},
        "subscription": {},
    }

    // برای هر فایل، یک set موقت برای جلوگیری از تکرار در همین اجرا
    fileSeen := make(map[string]map[string]bool)

    for cfg, entry := range entries {
        src := entry.Source
        if src != "telegram" && src != "subscription" {
            continue
        }
        // فقط کانفیگ‌های جدید از آخرین آرشیو به بعد
        if lastArch > 0 && entry.Timestamp < lastArch {
            continue
        }
        proto := entry.Protocol
        if proto == "" {
            proto = parser.DetectProtocol(cfg)
        }
        fileName := getFileNameForProtocol(proto)
        subDir := "📡 telegram"
        if src == "subscription" {
            subDir = "🔗 subscription"
        }
        fullPath := filepath.Join("📦 all_configs", subDir, fileName)

        if fileSeen[fullPath] == nil {
            fileSeen[fullPath] = make(map[string]bool)
        }
        if fileSeen[fullPath][cfg] {
            continue
        }
        fileSeen[fullPath][cfg] = true

        sourceFiles[src][fileName] = append(sourceFiles[src][fileName], cfg)
    }

    // نوشتن فایل‌ها (append به فایل‌های موجود)
    for src, fileMap := range sourceFiles {
        subDir := "📡 telegram"
        if src == "subscription" {
            subDir = "🔗 subscription"
        }
        for fname, configs := range fileMap {
            if len(configs) == 0 {
                continue
            }
            if sortOutput {
                sort.Slice(configs, func(i, j int) bool {
                    return entries[configs[i]].Timestamp > entries[configs[j]].Timestamp
                })
            }
            appendToFile(filepath.Join("📦 all_configs", subDir), fname, configs)
        }
    }
}

// getFileNameForProtocol بازگرداندن نام فایل (بدون پسوند) برای پروتکل
func getFileNameForProtocol(proto string) string {
    switch proto {
    case "vmess":
        return "📦 VMess.txt"
    case "vless":
        return "🕳️ VLess.txt"
    case "trojan":
        return "🐴 Trojan.txt"
    case "ss":
        return "🐍 Shadowsocks.txt"
    case "ssr":
        return "🔄 SSR.txt"
    case "hysteria2":
        return "⚡ Hysteria2.txt"
    case "tuic":
        return "🧩 Tuic.txt"
    case "wireguard":
        return "🔒 WireGuard.txt"
    case "warp":
        return "🌌 WARP.txt"
    case "mtproto":
        return "📱 MTProto Proxy.txt"
    case "telegram_socks":
        return "🧦 SOCKS5 Proxy.txt"
    case "socks5", "socks":
        return "🧦 SOCKS.txt"
    case "http":
        return "🌐 HTTP.txt"
    case "https":
        return "🔒 HTTPS.txt"
    case "argo":
        return "☁️ Argo.txt"
    case "slipnet":
        return "🕸️ Slipnet.txt"
    case "invizible":
        return "🛡️ Invizible Pro.txt"
    default:
        return "📄 all_protocols.txt"
    }
}

// appendToFile افزودن کانفیگ‌ها به انتهای فایل (ایجاد فایل در صورت نبود)
// تکراری‌های داخل فایل را هم حذف می‌کند (با خواندن محتوای قبلی)
func appendToFile(dir, filename string, configs []string) {
    path := filepath.Join(dir, filename)
    // اطمینان از وجود پوشه
    if err := os.MkdirAll(dir, 0755); err != nil {
        gologger.Error().Msgf("Cannot create directory %s: %v", dir, err)
        return
    }
    // خواندن محتوای قبلی برای حذف تکراری
    existing := make(map[string]bool)
    if data, err := os.ReadFile(path); err == nil {
        for _, line := range strings.Split(string(data), "\n") {
            line = strings.TrimSpace(line)
            if line != "" {
                existing[line] = true
            }
        }
    }
    // افزودن کانفیگ‌های جدید
    f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        gologger.Error().Msgf("Cannot open file %s: %v", path, err)
        return
    }
    defer f.Close()
    added := 0
    for _, cfg := range configs {
        if !existing[cfg] {
            if _, err := f.WriteString(cfg + "\n"); err != nil {
                gologger.Warning().Msgf("Failed to write to %s: %v", path, err)
            } else {
                added++
                existing[cfg] = true
            }
        }
    }
    if added > 0 {
        gologger.Info().Msgf("Added %d configs to %s", added, path)
    }
}
