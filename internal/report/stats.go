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

// GenerateStats آمار کامل را در یک فایل collector_stats.md تولید می‌کند (فایل متنی جداگانه حذف شده)
func GenerateStats(c *cache.Cache) {
    entries := c.GetAll()

    // آمار پایه
    total := len(entries)
    var tgCount, subCount int
    protoCount := make(map[string]int)
    channelCount := make(map[string]int)

    for _, e := range entries {
        if e.Source == "telegram" {
            tgCount++
            if e.Channel != "" {
                channelCount[e.Channel]++
            }
        } else if e.Source == "subscription" {
            subCount++
        }
        protoCount[e.Protocol]++
    }

    // یافتن جدیدترین و قدیمی‌ترین کانفیگ
    var newest, oldest int64
    first := true
    for _, e := range entries {
        if first {
            newest, oldest = e.Timestamp, e.Timestamp
            first = false
        } else {
            if e.Timestamp > newest {
                newest = e.Timestamp
            }
            if e.Timestamp < oldest {
                oldest = e.Timestamp
            }
        }
    }

    var sb strings.Builder
    sb.WriteString("# 📊 گزارش آماری پیشرفته جمع‌آوری‌کننده کانفیگ\n\n")
    sb.WriteString(fmt.Sprintf("**تاریخ گزارش:** `%s`\n\n", time.Now().Format("2006-01-02 15:04:05")))

    // خلاصه کلی با کارت
    sb.WriteString("## 📌 خلاصه کلی\n\n")
    sb.WriteString("| معیار | مقدار |\n|-------|-------|\n")
    sb.WriteString(fmt.Sprintf("| **کل کانفیگ‌های کش شده** | `%d` |\n", total))
    sb.WriteString(fmt.Sprintf("| **کانفیگ از تلگرام** | `%d` |\n", tgCount))
    sb.WriteString(fmt.Sprintf("| **کانفیگ از ساب‌لینک** | `%d` |\n", subCount))
    if !first {
        sb.WriteString(fmt.Sprintf("| **بازه زمانی کش** | `%s` تا `%s` |\n",
            time.Unix(oldest, 0).Format("2006-01-02"), time.Unix(newest, 0).Format("2006-01-02")))
    }
    sb.WriteString("\n")

    // تفکیک پروتکل با نمودار متنی
    sb.WriteString("## 📡 تفکیک پروتکل (تعداد)\n\n")
    if len(protoCount) == 0 {
        sb.WriteString("_هیچ پروتکلی یافت نشد._\n\n")
    } else {
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
    }

    // کانال‌های برتر تلگرام
    if len(channelCount) > 0 {
        sb.WriteString("## 🏆 برترین کانال‌های تلگرام (بر اساس تعداد کانفیگ)\n\n")
        type kv struct {
            ch string
            c  int
        }
        var sorted []kv
        for ch, c := range channelCount {
            sorted = append(sorted, kv{ch, c})
        }
        sort.Slice(sorted, func(i, j int) bool { return sorted[i].c > sorted[j].c })
        sb.WriteString("| رتبه | کانال | تعداد کانفیگ |\n")
        sb.WriteString("|------|-------|--------------|\n")
        for i, kv := range sorted {
            if i >= 20 {
                sb.WriteString(fmt.Sprintf("| ... | و `%d` کانال دیگر | ... |\n", len(sorted)-20))
                break
            }
            sb.WriteString(fmt.Sprintf("| %d | `%s` | `%d` |\n", i+1, kv.ch, kv.c))
        }
        sb.WriteString("\n")
    }

    // لینک‌های مفید
    sb.WriteString("## 🔗 لینک‌های مفید\n\n")
    sb.WriteString(fmt.Sprintf("- [مشاهده لینک‌های دانلود مستقیم](%s/reports/links.md)\n", detectRawBaseURL()))
    sb.WriteString(fmt.Sprintf("- [مشاهده آخرین اجرای اکشن](https://github.com/%s/actions)\n", os.Getenv("GITHUB_REPOSITORY")))

    sb.WriteString("\n---\n✅ این گزارش به صورت خودکار هر بار تولید می‌شود.\n")

    os.MkdirAll("reports", 0755)
    os.WriteFile("reports/collector_stats.md", []byte(sb.String()), 0644)
    // حذف collector_stats.txt
    os.Remove("reports/collector_stats.txt")
}
