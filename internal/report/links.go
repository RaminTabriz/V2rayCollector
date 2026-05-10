package report

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/RaminTabriz/V2rayCollector/internal/cache"
)

// GenerateLinks تولید فایل reports/links.md با لینک‌های دانلود مستقیم فایل‌های خروجی
func GenerateLinks(c *cache.Cache) {
    // تشخیص base URL برای لینک‌های خام گیت‌هاب
    baseURL := detectBaseURL()

    var sb strings.Builder
    sb.WriteString("# 🔗 لینک‌های فایل‌های کانفیگ\n\n")
    sb.WriteString(fmt.Sprintf("**آخرین به‌روزرسانی:** `%s`\n\n", time.Now().Format("2006-01-02 15:04:05")))
    sb.WriteString("---\n\n## 📊 گزارش آماری کامل\n\n")
    sb.WriteString(fmt.Sprintf("- [📈 collector_stats.md](%s/reports/collector_stats.md)\n", baseURL))
    sb.WriteString(fmt.Sprintf("- [📄 collector_stats.txt](%s/reports/collector_stats.txt)\n\n", baseURL))
    sb.WriteString("---\n\n")

    // تابع کمکی برای استخراج اطلاعات فایل
    getFileInfo := func(path string) (lines int, modTime time.Time) {
        info, err := os.Stat(path)
        if err != nil {
            return 0, time.Time{}
        }
        data, err := os.ReadFile(path)
        if err != nil {
            return 0, info.ModTime()
        }
        lines = strings.Count(string(data), "\n")
        if lines == 0 && len(data) > 0 {
            lines = 1
        }
        return lines, info.ModTime()
    }

    statusEmoji := func(lines int) string {
        if lines == 0 {
            return "🔴"
        }
        return "🟢"
    }

    // پوشه all_configs/subscription
    sb.WriteString("## 📁 all_configs/subscription\n\n")
    sb.WriteString("| وضعیت | نام فایل | تعداد کانفیگ | آخرین به‌روزرسانی | لینک خام |\n")
    sb.WriteString("|-------|----------|--------------|-------------------|----------|\n")
    files, _ := filepath.Glob("📦 all_configs/🔗 subscription/*.txt")
    for _, f := range files {
        name := filepath.Base(f)
        lines, mod := getFileInfo(f)
        status := statusEmoji(lines)
        url := fmt.Sprintf("%s/%s", baseURL, strings.ReplaceAll(f, "\\", "/"))
        timeStr := mod.Format("2006-01-02 15:04:05")
        if mod.IsZero() {
            timeStr = "نامشخص"
        }
        sb.WriteString(fmt.Sprintf("| %s | `%s` | `%d` | `%s` | [دانلود](%s) |\n", status, name, lines, timeStr, url))
    }
    sb.WriteString("\n")

    // پوشه all_configs/telegram
    sb.WriteString("## 📁 all_configs/telegram\n\n")
    sb.WriteString("| وضعیت | نام فایل | تعداد کانفیگ | آخرین به‌روزرسانی | لینک خام |\n")
    sb.WriteString("|-------|----------|--------------|-------------------|----------|\n")
    files, _ = filepath.Glob("📦 all_configs/📡 telegram/*.txt")
    for _, f := range files {
        name := filepath.Base(f)
        lines, mod := getFileInfo(f)
        status := statusEmoji(lines)
        url := fmt.Sprintf("%s/%s", baseURL, strings.ReplaceAll(f, "\\", "/"))
        timeStr := mod.Format("2006-01-02 15:04:05")
        if mod.IsZero() {
            timeStr = "نامشخص"
        }
        sb.WriteString(fmt.Sprintf("| %s | `%s` | `%d` | `%s` | [دانلود](%s) |\n", status, name, lines, timeStr, url))
    }
    sb.WriteString("\n")

    // سایر پوشه‌های اصلی
    for _, folder := range []string{"📡 telegram", "🔗 subscription", "🌍 mixed"} {
        sb.WriteString(fmt.Sprintf("## 📁 %s\n\n", folder))
        sb.WriteString("| وضعیت | نام فایل | تعداد کانفیگ | آخرین به‌روزرسانی | لینک خام |\n")
        sb.WriteString("|-------|----------|--------------|-------------------|----------|\n")
        files, _ = filepath.Glob(fmt.Sprintf("%s/*.txt", folder))
        for _, f := range files {
            name := filepath.Base(f)
            lines, mod := getFileInfo(f)
            status := statusEmoji(lines)
            url := fmt.Sprintf("%s/%s", baseURL, strings.ReplaceAll(f, "\\", "/"))
            timeStr := mod.Format("2006-01-02 15:04:05")
            if mod.IsZero() {
                timeStr = "نامشخص"
            }
            sb.WriteString(fmt.Sprintf("| %s | `%s` | `%d` | `%s` | [دانلود](%s) |\n", status, name, lines, timeStr, url))
        }
        sb.WriteString("\n")
    }

    os.MkdirAll("reports", 0755)
    os.WriteFile("reports/links.md", []byte(sb.String()), 0644)
}

// detectBaseURL تشخیص base URL برای فایل‌های خام گیت‌هاب (از محیط یا پیش‌فرض)
func detectBaseURL() string {
    // در GitHub Actions می‌توان از متغیرهای محیطی استفاده کرد
    repo := os.Getenv("GITHUB_REPOSITORY")
    if repo == "" {
        repo = "RaminTabriz/V2rayCollector"
    }
    branch := os.Getenv("GITHUB_REF_NAME")
    if branch == "" {
        branch = "main"
    }
    return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s", repo, branch)
}
