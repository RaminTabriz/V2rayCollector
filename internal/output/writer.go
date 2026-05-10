package output

import (
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"

    "github.com/RaminTabriz/V2rayCollector/internal/cache"
    "github.com/RaminTabriz/V2rayCollector/internal/parser"
)

// WriteTelegramFiles نوشتن کانفیگ‌های تلگرام در پوشه 📡 telegram
// هر کانال در زیرپوشه خودش ذخیره می‌شود (با نام کانال) و فایل‌ها بر اساس پروتکل جدا می‌شوند
func WriteTelegramFiles(c *cache.Cache, sortOutput bool) {
    entries := c.GetAll()
    byChannel := make(map[string]map[string][]string) // channel -> protocol -> []config

    for cfg, e := range entries {
        if e.Source != "telegram" || e.Channel == "" {
            continue
        }
        proto := e.Protocol
        if proto == "" {
            proto = parser.DetectProtocol(cfg)
        }
        if byChannel[e.Channel] == nil {
            byChannel[e.Channel] = make(map[string][]string)
        }
        byChannel[e.Channel][proto] = append(byChannel[e.Channel][proto], cfg)
    }

    for channel, protoMap := range byChannel {
        channelDir := filepath.Join("📡 telegram", channel)
        if err := os.MkdirAll(channelDir, 0755); err != nil {
            fmt.Printf("Error creating dir %s: %v\n", channelDir, err)
            continue
        }
        for proto, configs := range protoMap {
            if sortOutput {
                sort.Slice(configs, func(i, j int) bool {
                    return entries[configs[i]].Timestamp > entries[configs[j]].Timestamp
                })
            }
            content := strings.Join(configs, "\n")
            if content == "" {
                continue
            }
            fileName := parser.GetProtocolFileName(proto) + ".txt"
            filePath := filepath.Join(channelDir, fileName)
            if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
                fmt.Printf("Error writing %s: %v\n", filePath, err)
            } else {
                fmt.Printf("✅ Wrote %d configs to %s\n", len(configs), filePath)
            }
        }
    }
}

// WriteSubscriptionFiles نوشتن کانفیگ‌های ساب‌لینک در پوشه 🔗 subscription
// فایل‌ها بر اساس پروتکل جدا می‌شوند
func WriteSubscriptionFiles(c *cache.Cache, sortOutput bool) {
    entries := c.GetAll()
    byProtocol := make(map[string][]string)

    for cfg, e := range entries {
        if e.Source != "subscription" {
            continue
        }
        proto := e.Protocol
        if proto == "" {
            proto = parser.DetectProtocol(cfg)
        }
        byProtocol[proto] = append(byProtocol[proto], cfg)
    }

    if err := os.MkdirAll("🔗 subscription", 0755); err != nil {
        fmt.Printf("Error creating dir: %v\n", err)
        return
    }

    for proto, configs := range byProtocol {
        if sortOutput {
            sort.Slice(configs, func(i, j int) bool {
                return entries[configs[i]].Timestamp > entries[configs[j]].Timestamp
            })
        }
        content := strings.Join(configs, "\n")
        if content == "" {
            continue
        }
        fileName := parser.GetProtocolFileName(proto) + ".txt"
        filePath := filepath.Join("🔗 subscription", fileName)
        if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
            fmt.Printf("Error writing %s: %v\n", filePath, err)
        } else {
            fmt.Printf("✅ Wrote %d configs to %s\n", len(configs), filePath)
        }
    }
}

// WriteMixedFiles نوشتن کانفیگ‌های با پروتکل نامشخص (mixed) در پوشه 🌍 mixed
func WriteMixedFiles(c *cache.Cache, sortOutput bool) {
    entries := c.GetAll()
    var mixed []string

    for cfg, e := range entries {
        proto := e.Protocol
        if proto == "" {
            proto = parser.DetectProtocol(cfg)
        }
        if proto == "mixed" {
            mixed = append(mixed, cfg)
        }
    }

    if len(mixed) == 0 {
        return
    }

    if sortOutput {
        sort.Slice(mixed, func(i, j int) bool {
            return entries[mixed[i]].Timestamp > entries[mixed[j]].Timestamp
        })
    }

    if err := os.MkdirAll("🌍 mixed", 0755); err != nil {
        fmt.Printf("Error creating dir: %v\n", err)
        return
    }
    filePath := filepath.Join("🌍 mixed", parser.GetProtocolFileName("mixed")+".txt")
    content := strings.Join(mixed, "\n")
    if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
        fmt.Printf("Error writing %s: %v\n", filePath, err)
    } else {
        fmt.Printf("✅ Wrote %d mixed configs to %s\n", len(mixed), filePath)
    }
}

// GenerateClashYAML تولید فایل clash-config.yaml (ساده)
func GenerateClashYAML(c *cache.Cache) {
    entries := c.GetAll()
    var proxies []string
    for cfg, e := range entries {
        if e.Source == "subscription" || e.Source == "telegram" {
            // تبدیل ساده به فرمت clash (فقط نمایش - قابل گسترش)
            proxies = append(proxies, fmt.Sprintf("  - name: \"%s_%d\"\n    type: %s\n    server: placeholder\n    port: 443\n", e.Protocol, e.Timestamp, e.Protocol))
        }
    }
    content := "proxies:\n" + strings.Join(proxies, "")
    os.WriteFile("clash-config.yaml", []byte(content), 0644)
    fmt.Println("✅ Generated clash-config.yaml (placeholder)")
}
