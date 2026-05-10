package report

import (
    "fmt"
    "os"
    "sort"
    "strings"
    "time"

    "github.com/RaminTabriz/V2rayCollector/internal/cache"
    "github.com/RaminTabriz/V2rayCollector/internal/parser"
)

// GenerateStats تولید گزارش آماری در فایل reports/collector_stats.md
func GenerateStats(c *cache.Cache) {
    entries := c.GetAll()

    totalConfigs := len(entries)
    var newCount, dupCount, insecureCount, telegramCount, subCount int
    protoCounts := make(map[string]int)
    channelStats := make(map[string]int)

    for _, e := range entries {
        if e.Source == "telegram" {
            telegramCount++
            if e.Channel != "" {
                channelStats[e.Channel]++
            }
        } else if e.Source == "subscription" {
            subCount++
        }
        protoCounts[e.Protocol]++
    }

    // آمار جدید، تکراری و ناامن در این اجرا قابل محاسبه نیست مگر اینکه در کش ذخیره شده باشند
    // برای سادگی از کش استفاده می‌کنیم (در حال حاضر فقط آمار کش را نشان می‌دهیم)

    var sb strings.Builder
    sb.WriteString("# 📊 گزارش آماری جمع‌آوری‌کننده کانفیگ\n\n")
    sb.WriteString(fmt.Sprintf("**زمان اجرا:** `%s`\n\n", time.Now().Format("2006-01-02 15:04:05")))

    sb.WriteString("## 📈 آمار کلی\n\n")
    sb.WriteString("| معیار | مقدار |\n|-------|-------|\n")
    sb.WriteString(fmt.Sprintf("| **کل کانفیگ‌های کش** | `%d` |\n", totalConfigs))
    sb.WriteString(fmt.Sprintf("| **کانفیگ از تلگرام** | `%d` |\n", telegramCount))
    sb.WriteString(fmt.Sprintf("| **کانفیگ از ساب‌لینک** | `%d` |\n\n", subCount))

    sb.WriteString("## 📡 تفکیک پروتکل‌ها\n\n")
    if len(protoCounts) == 0 {
        sb.WriteString("_هیچ پروتکلی یافت نشد._\n\n")
    } else {
        sb.WriteString("| پروتکل | تعداد |\n|--------|-------|\n")
        type kv struct {
            p string
            c int
        }
        var sorted []kv
        for p, c := range protoCounts {
            sorted = append(sorted, kv{p, c})
        }
        sort.Slice(sorted, func(i, j int) bool { return sorted[i].c > sorted[j].c })
        for _, kv := range sorted {
            icon := parser.GetProtocolIcon(kv.p)
            sb.WriteString(fmt.Sprintf("| %s **%s** | `%d` |\n", icon, kv.p, kv.c))
        }
        sb.WriteString("\n")
    }

    if len(channelStats) > 0 {
        sb.WriteString("## 🗂️ کانال‌های تلگرام (به تفکیک تعداد کانفیگ)\n\n")
        type kv struct {
            ch string
            c  int
        }
        var sorted []kv
        for ch, c := range channelStats {
            sorted = append(sorted, kv{ch, c})
        }
        sort.Slice(sorted, func(i, j int) bool { return sorted[i].c > sorted[j].c })
        sb.WriteString("| کانال | تعداد کانفیگ |\n|-------|--------------|\n")
        for i, kv := range sorted {
            if i >= 20 {
                sb.WriteString(fmt.Sprintf("| ... و `%d` کانال دیگر | ... |\n", len(sorted)-20))
                break
            }
            sb.WriteString(fmt.Sprintf("| `%s` | `%d` |\n", kv.ch, kv.c))
        }
        sb.WriteString("\n")
    }

    sb.WriteString("---\n✅ گزارش توسط V2rayCollector تولید شده است.\n")
    os.MkdirAll("reports", 0755)
    os.WriteFile("reports/collector_stats.md", []byte(sb.String()), 0644)
    // همچنین یک نسخه متنی ساده
    os.WriteFile("reports/collector_stats.txt", []byte(sb.String()), 0644)
}
