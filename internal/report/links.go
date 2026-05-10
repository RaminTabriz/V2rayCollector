package report

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"
)

// GenerateLinks فایل links.md را با جدول کامل لینک‌های خام و اطلاعات دقیق تولید می‌کند
func GenerateLinks() {
    baseRawURL := detectRawBaseURL()
    now := time.Now()

    var sb strings.Builder
    sb.WriteString("# 🔗 لینک‌های مستقیم (Raw) فایل‌های کانفیگ\n\n")
    sb.WriteString(fmt.Sprintf("**آخرین به‌روزرسانی:** `%s`\n\n", now.Format("2006-01-02 15:04:05")))
    sb.WriteString("> این لینک‌ها برای استفاده در نرم‌افزارهایی مانند **V2RayNG**, **Clash**, **Sing-box** قابل استفاده هستند.\n\n")
    sb.WriteString("---\n\n")

    // ========================== all_configs/subscription ==========================
    sb.WriteString("## 📦 all_configs/subscription (ساب‌لینک‌های تجمیعی)\n\n")
    sb.WriteString("| نام فایل | حجم | تعداد کانفیگ | آخرین تغییر | لینک خام (Raw) |\n")
    sb.WriteString("|----------|------|--------------|--------------|----------------|\n")
    files, _ := filepath.Glob("📦 all_configs/🔗 subscription/*.txt")
    for _, f := range files {
        name := filepath.Base(f)
        size, lines, mod := fileStats(f)
        url := fmt.Sprintf("%s/%s", baseRawURL, strings.ReplaceAll(f, "\\", "/"))
        sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%d` | `%s` | [دانلود](%s) |\n",
            name, size, lines, mod.Format("2006-01-02 15:04:05"), url))
    }
    sb.WriteString("\n")

    // ========================== all_configs/telegram ==========================
    sb.WriteString("## 📡 all_configs/telegram (کانفیگ‌های روزانه تلگرام)\n\n")
    sb.WriteString("| نام فایل | حجم | تعداد کانفیگ | آخرین تغییر | لینک خام (Raw) |\n")
    sb.WriteString("|----------|------|--------------|--------------|----------------|\n")
    files, _ = filepath.Glob("📦 all_configs/📡 telegram/*.txt")
    for _, f := range files {
        name := filepath.Base(f)
        size, lines, mod := fileStats(f)
        url := fmt.Sprintf("%s/%s", baseRawURL, strings.ReplaceAll(f, "\\", "/"))
        sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%d` | `%s` | [دانلود](%s) |\n",
            name, size, lines, mod.Format("2006-01-02 15:04:05"), url))
    }
    sb.WriteString("\n")

    // ========================== دسته‌بندی اصلی ==========================
    for _, folder := range []string{"📡 telegram", "🔗 subscription", "🌍 mixed"} {
        sb.WriteString(fmt.Sprintf("## 🗂️ %s (خروجی اصلی)\n\n", folder))
        sb.WriteString("| نام فایل | حجم | تعداد کانفیگ | آخرین تغییر | لینک خام (Raw) |\n")
        sb.WriteString("|----------|------|--------------|--------------|----------------|\n")
        files, _ = filepath.Glob(fmt.Sprintf("%s/*.txt", folder))
        for _, f := range files {
            name := filepath.Base(f)
            size, lines, mod := fileStats(f)
            url := fmt.Sprintf("%s/%s", baseRawURL, strings.ReplaceAll(f, "\\", "/"))
            sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%d` | `%s` | [دانلود](%s) |\n",
                name, size, lines, mod.Format("2006-01-02 15:04:05"), url))
        }
        sb.WriteString("\n")
    }

    // ========================== آرشیو روزانه (فقط تاریخ امروز) ==========================
    today := now.Format("2006-01-02")
    archiveDir := fmt.Sprintf("🗄️ daily_archive/%s/📦 all_configs", today)
    if _, err := os.Stat(archiveDir); err == nil {
        sb.WriteString("## 📦 آرشیو امروز (فقط برای امروز)\n\n")
        for _, sub := range []string{"🔗 subscription", "📡 telegram"} {
            subPath := filepath.Join(archiveDir, sub)
            files, _ = filepath.Glob(fmt.Sprintf("%s/*.txt", subPath))
            if len(files) == 0 {
                continue
            }
            sb.WriteString(fmt.Sprintf("### %s\n\n", sub))
            sb.WriteString("| نام فایل | حجم | تعداد کانفیگ | لینک خام (Raw) |\n")
            sb.WriteString("|----------|------|--------------|----------------|\n")
            for _, f := range files {
                name := filepath.Base(f)
                size, lines, _ := fileStats(f)
                url := fmt.Sprintf("%s/%s", baseRawURL, strings.ReplaceAll(f, "\\", "/"))
                sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%d` | [دانلود](%s) |\n",
                    name, size, lines, url))
            }
            sb.WriteString("\n")
        }
    }

    // ========================== گزارش آماری ==========================
    sb.WriteString("## 📊 گزارش آماری\n\n")
    sb.WriteString(fmt.Sprintf("- [مشاهده گزارش کامل آمار](%s/reports/collector_stats.md)\n", baseRawURL))
    sb.WriteString("\n---\n✅ این فایل به صورت خودکار هر بار تولید می‌شود.\n")

    os.MkdirAll("reports", 0755)
    os.WriteFile("reports/links.md", []byte(sb.String()), 0644)
}

// detectRawBaseURL آدرس base را برای لینک‌های خام گیت‌هاب تشخیص می‌دهد
func detectRawBaseURL() string {
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

// fileStats اطلاعات یک فایل را برمی‌گرداند: حجم (human readable), تعداد خطوط, زمان آخرین تغییر
func fileStats(path string) (sizeStr string, lines int, modTime time.Time) {
    info, err := os.Stat(path)
    if err != nil {
        return "0 B", 0, time.Time{}
    }
    size := info.Size()
    switch {
    case size < 1024:
        sizeStr = fmt.Sprintf("%d B", size)
    case size < 1024*1024:
        sizeStr = fmt.Sprintf("%.1f KB", float64(size)/1024)
    default:
        sizeStr = fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
    }
    data, err := os.ReadFile(path)
    if err == nil {
        lines = strings.Count(string(data), "\n")
        if lines == 0 && len(data) > 0 {
            lines = 1
        }
    }
    modTime = info.ModTime()
    return
}
