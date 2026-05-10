package report

import (
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/RaminTabriz/V2rayCollector/internal/cache"
    "github.com/RaminTabriz/V2rayCollector/internal/parser"
    "github.com/projectdiscovery/gologger"
)

type TrafficType struct {
    TCP  int
    WS   int
    gRPC int
    HTTP int
    QUIC int
    Other int
}

type HotServer struct {
    Address string
    Count   int
}

func GenerateStats(c *cache.Cache, subSuccess, subFailed int, activeChannels, deadChannels int) {
    entries := c.GetAll()
    total := len(entries)

    // آمار اولیه
    var tgCount, subCount int
    protoCount := make(map[string]int)
    channelCount := make(map[string]int)
    dailyCount := make(map[string]int) // key: YYYY-MM-DD
    traffic := TrafficType{}
    serverMap := make(map[string]int)

    now := time.Now()
    today := now.Format("2006-01-02")
    yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
    weekAgo := now.AddDate(0, 0, -7)
    monthAgo := now.AddDate(0, -1, 0)

    var todayCount, yesterdayCount, weekCount, monthCount int

    for _, e := range entries {
        // منبع
        if e.Source == "telegram" {
            tgCount++
            if e.Channel != "" {
                channelCount[e.Channel]++
            }
        } else if e.Source == "subscription" {
            subCount++
        }

        // پروتکل
        protoCount[e.Protocol]++

        // تاریخ (روز)
        date := time.Unix(e.Timestamp, 0).Format("2006-01-02")
        dailyCount[date]++

        // فیلترهای زمانی
        if date == today {
            todayCount++
        }
        if date == yesterday {
            yesterdayCount++
        }
        t := time.Unix(e.Timestamp, 0)
        if t.After(weekAgo) {
            weekCount++
        }
        if t.After(monthAgo) {
            monthCount++
        }

        // نوع ترافیک (از کانفیگ اصلی)
        trafficType := detectTrafficType(e.Original, e.Protocol)
        switch trafficType {
        case "tcp":
            traffic.TCP++
        case "ws":
            traffic.WS++
        case "grpc":
            traffic.gRPC++
        case "http":
            traffic.HTTP++
        case "quic":
            traffic.QUIC++
        default:
            traffic.Other++
        }

        // سرور پرتکرار
        server := extractServerAddress(e.Original, e.Protocol)
        if server != "" {
            serverMap[server]++
        }
    }

    // ۱۰ سرور پرتکرار
    var hotServers []HotServer
    for addr, cnt := range serverMap {
        hotServers = append(hotServers, HotServer{addr, cnt})
    }
    sort.Slice(hotServers, func(i, j int) bool { return hotServers[i].Count > hotServers[j].Count })
    if len(hotServers) > 10 {
        hotServers = hotServers[:10]
    }

    // زمان‌بندی ۷ روز اخیر
    timeline := make([]string, 7)
    for i := 6; i >= 0; i-- {
        d := now.AddDate(0, 0, -i).Format("2006-01-02")
        cnt := dailyCount[d]
        timeline[i] = fmt.Sprintf("%s: %d", d, cnt)
    }

    // ========== ساخت گزارش ==========
    var sb strings.Builder
    sb.WriteString("# 📊 گزارش آماری پیشرفته V2rayCollector\n\n")
    sb.WriteString(fmt.Sprintf("**تاریخ گزارش:** `%s`\n\n", now.Format("2006-01-02 15:04:05")))

    // 1. خلاصه کلی
    sb.WriteString("## 📌 خلاصه کلی\n\n")
    sb.WriteString("| معیار | مقدار |\n|-------|-------|\n")
    sb.WriteString(fmt.Sprintf("| **کل کانفیگ‌های کش شده** | `%d` |\n", total))
    sb.WriteString(fmt.Sprintf("| **کانفیگ از تلگرام** | `%d` |\n", tgCount))
    sb.WriteString(fmt.Sprintf("| **کانفیگ از ساب‌لینک** | `%d` |\n", subCount))
    sb.WriteString(fmt.Sprintf("| **کانال‌های فعال** | `%d` |\n", activeChannels))
    sb.WriteString(fmt.Sprintf("| **کانال‌های غیرفعال** | `%d` |\n", deadChannels))
    sb.WriteString("\n")

    // 2. فیلتر تاریخ
    sb.WriteString("## 📅 آمار زمانی\n\n")
    sb.WriteString("| بازه | تعداد کانفیگ |\n|------|-------------|\n")
    sb.WriteString(fmt.Sprintf("| **امروز** | `%d` |\n", todayCount))
    sb.WriteString(fmt.Sprintf("| **دیروز** | `%d` |\n", yesterdayCount))
    sb.WriteString(fmt.Sprintf("| **۷ روز اخیر** | `%d` |\n", weekCount))
    sb.WriteString(fmt.Sprintf("| **۳۰ روز اخیر** | `%d` |\n", monthCount))
    sb.WriteString("\n")

    // 3. نمودار رشد ۷ روزه
    sb.WriteString("## 📈 نمودار رشد (۷ روز اخیر)\n\n")
    maxCount := 0
    for _, d := range timeline {
        parts := strings.Split(d, ": ")
        if len(parts) == 2 {
            cnt, _ := strconv.Atoi(parts[1])
            if cnt > maxCount {
                maxCount = cnt
            }
        }
    }
    for _, d := range timeline {
        parts := strings.Split(d, ": ")
        if len(parts) == 2 {
            datePart := parts[0]
            cnt, _ := strconv.Atoi(parts[1])
            barLen := 0
            if maxCount > 0 {
                barLen = int(float64(cnt) / float64(maxCount) * 40)
            }
            bar := strings.Repeat("█", barLen)
            sb.WriteString(fmt.Sprintf("`%s` : %s `(%d)`\n", datePart, bar, cnt))
        }
    }
    sb.WriteString("\n")

    // 4. تفکیک پروتکل با نوار پیشرفت
    sb.WriteString("## 📡 تفکیک پروتکل\n\n")
    type kv struct {
        p string
        c int
    }
    var sorted []kv
    for p, c := range protoCount {
        sorted = append(sorted, kv{p, c})
    }
    sort.Slice(sorted, func(i, j int) bool { return sorted[i].c > sorted[j].c })
    sb.WriteString("| پروتکل | تعداد | درصد | نوار پیشرفت |\n")
    sb.WriteString("|--------|-------|------|-------------|\n")
    for _, kv := range sorted {
        percent := float64(kv.c) / float64(total) * 100
        barLen := int(percent / 2)
        bar := strings.Repeat("█", barLen) + strings.Repeat("░", 50-barLen)
        icon := parser.GetProtocolIcon(kv.p)
        sb.WriteString(fmt.Sprintf("| %s **%s** | `%d` | `%.1f%%` | `%s` |\n", icon, kv.p, kv.c, percent, bar))
    }
    sb.WriteString("\n")

    // 5. آمار نوع ترافیک
    sb.WriteString("## 🌐 تفکیک نوع ترافیک\n\n")
    sb.WriteString("| نوع | تعداد |\n|------|-------|\n")
    sb.WriteString(fmt.Sprintf("| TCP | `%d` |\n", traffic.TCP))
    sb.WriteString(fmt.Sprintf("| WebSocket (WS) | `%d` |\n", traffic.WS))
    sb.WriteString(fmt.Sprintf("| gRPC | `%d` |\n", traffic.gRPC))
    sb.WriteString(fmt.Sprintf("| HTTP | `%d` |\n", traffic.HTTP))
    sb.WriteString(fmt.Sprintf("| QUIC | `%d` |\n", traffic.QUIC))
    sb.WriteString(fmt.Sprintf("| سایر | `%d` |\n", traffic.Other))
    sb.WriteString("\n")

    // 6. برترین کانال‌ها
    if len(channelCount) > 0 {
        sb.WriteString("## 🏆 برترین کانال‌های تلگرام\n\n")
        sortedCh := []kv{}
        for ch, c := range channelCount {
            sortedCh = append(sortedCh, kv{ch, c})
        }
        sort.Slice(sortedCh, func(i, j int) bool { return sortedCh[i].c > sortedCh[j].c })
        sb.WriteString("| رتبه | کانال | تعداد کانفیگ |\n")
        sb.WriteString("|------|-------|--------------|\n")
        for i, kv := range sortedCh {
            if i >= 20 {
                sb.WriteString(fmt.Sprintf("| ... | و `%d` کانال دیگر | ... |\n", len(sortedCh)-20))
                break
            }
            sb.WriteString(fmt.Sprintf("| %d | `%s` | `%d` |\n", i+1, kv.ch, kv.c))
        }
        sb.WriteString("\n")
    }

    // 7. برترین ساب‌لینک‌ها (با استفاده از آمار موفق/ناموفق - فعلاً placeholder)
    sb.WriteString("## 🔗 وضعیت ساب‌لینک‌ها (آخرین اجرا)\n\n")
    sb.WriteString(fmt.Sprintf("| موفق | ناموفق |成功率 |\n"))
    sb.WriteString(fmt.Sprintf("|------|-------|------|\n"))
    totalSub := subSuccess + subFailed
    successRate := 0.0
    if totalSub > 0 {
        successRate = float64(subSuccess) / float64(totalSub) * 100
    }
    sb.WriteString(fmt.Sprintf("| `%d` | `%d` | `%.1f%%` |\n", subSuccess, subFailed, successRate))
    sb.WriteString("\n")

    // 8. فضای ذخیره‌سازی
    totalSize := getTotalSize("📦 all_configs") + getTotalSize("🗄️ daily_archive")
    sb.WriteString("## 💾 فضای ذخیره‌سازی اشغال شده\n\n")
    sb.WriteString(fmt.Sprintf("| **کل فضای اشغالی** | `%s` |\n", formatBytes(totalSize)))
    sb.WriteString("\n")

    // 9. جدول ۱۰ سرور پرتکرار
    sb.WriteString("## 🔥 ۱۰ سرور پرتکرار (Hot Servers)\n\n")
    sb.WriteString("| رتبه | آدرس سرور | تعداد تکرار |\n")
    sb.WriteString("|------|-----------|-------------|\n")
    for i, s := range hotServers {
        sb.WriteString(fmt.Sprintf("| %d | `%s` | `%d` |\n", i+1, s.Address, s.Count))
    }
    sb.WriteString("\n")

    // 10. جستجو و فیلتر (قسمت قابل جمع‌شدن)
    sb.WriteString("<details>\n<summary>🔍 نمایش جزئیات پروتکل‌ها (قابل جستجو)</summary>\n\n")
    sb.WriteString("```\n")
    for _, kv := range sorted {
        sb.WriteString(fmt.Sprintf("%s: %d\n", kv.p, kv.c))
    }
    sb.WriteString("```\n</details>\n\n")

    sb.WriteString("---\n✅ این گزارش به صورت خودکار هر بار تولید می‌شود.\n")

    os.MkdirAll("reports", 0755)
    os.WriteFile("reports/collector_stats.md", []byte(sb.String()), 0644)
    os.Remove("reports/collector_stats.txt")
}

// Helper functions
func detectTrafficType(cfg, proto string) string {
    lower := strings.ToLower(cfg)
    if strings.Contains(lower, "type=ws") || strings.Contains(lower, "network=ws") {
        return "ws"
    }
    if strings.Contains(lower, "type=grpc") || strings.Contains(lower, "network=grpc") {
        return "grpc"
    }
    if strings.Contains(lower, "type=http") || strings.Contains(lower, "headerType=http") {
        return "http"
    }
    if strings.Contains(lower, "type=quic") || strings.Contains(lower, "quic") {
        return "quic"
    }
    if strings.Contains(lower, "type=tcp") || strings.Contains(lower, "network=tcp") {
        return "tcp"
    }
    return "other"
}

func extractServerAddress(cfg, proto string) string {
    // استخراج ساده IP/domain
    switch proto {
    case "vmess":
        // if json encoded
        if strings.HasPrefix(cfg, "vmess://") {
            // decode and get "add"
            // simplified: just find after @ or after //
            // better: use regex
        }
    case "vless", "trojan", "ss", "ssr", "hysteria2", "tuic":
        // format: proto://user:pass@host:port
        if idx := strings.Index(cfg, "@"); idx != -1 {
            after := cfg[idx+1:]
            if colon := strings.Index(after, ":"); colon != -1 {
                return after[:colon]
            }
            if slash := strings.Index(after, "/"); slash != -1 {
                return after[:slash]
            }
            return after
        }
    }
    // fallback: simple regex for IP or domain
    re := regexp.MustCompile(`(?i)(?:[a-z0-9-]+\.)+[a-z]{2,}|\d+\.\d+\.\d+\.\d+`)
    match := re.FindString(cfg)
    if match != "" {
        return match
    }
    return ""
}

func getTotalSize(dir string) int64 {
    var size int64
    filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
        if err == nil && !info.IsDir() {
            size += info.Size()
        }
        return nil
    })
    return size
}

func formatBytes(b int64) string {
    const unit = 1024
    if b < unit {
        return fmt.Sprintf("%d B", b)
    }
    div, exp := int64(unit), 0
    for n := b / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
