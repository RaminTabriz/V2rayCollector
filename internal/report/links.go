package report

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"
)

func GenerateLinks() {
    baseRawURL := detectRawBaseURL()
    now := time.Now()

    var sb strings.Builder
    sb.WriteString("# 🔗 لینک‌های مستقیم (Raw) فایل‌های کانفیگ\n\n")
    sb.WriteString(fmt.Sprintf("**آخرین به‌روزرسانی:** `%s`\n\n", now.Format("2006-01-02 15:04:05")))
    sb.WriteString("> این لینک‌ها برای استفاده در نرم‌افزارهایی مانند **V2RayNG**, **Clash**, **Sing-box** قابل استفاده هستند.\n\n")
    sb.WriteString("---\n\n")

    // Helper: get valid config lines count
    countValidConfigs := func(path string) int {
        data, err := os.ReadFile(path)
        if err != nil {
            return 0
        }
        lines := strings.Split(string(data), "\n")
        valid := 0
        for _, line := range lines {
            line = strings.TrimSpace(line)
            if line != "" && (strings.Contains(line, "://") || strings.Contains(line, "-----BEGIN ARGO")) {
                valid++
            }
        }
        return valid
    }

    // all_configs/subscription
    sb.WriteString("## 📦 all_configs/subscription (ساب‌لینک‌های تجمیعی)\n\n")
    sb.WriteString("| نام فایل | حجم | تعداد خطوط | تعداد کل پروکسی | آخرین تغییر | لینک خام (Raw) |\n")
    sb.WriteString("|----------|------|------------|----------------|--------------|----------------|\n")
    files, _ := filepath.Glob("📦 all_configs/🔗 subscription/*.txt")
    for _, f := range files {
        name := filepath.Base(f)
        size, lines, mod := fileStats(f)
        validCount := countValidConfigs(f)
        url := fmt.Sprintf("%s/%s", baseRawURL, strings.ReplaceAll(f, "\\", "/"))
        sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%d` | `%d` | `%s` | [دانلود](%s) |\n",
            name, size, lines, validCount, mod.Format("2006-01-02 15:04:05"), url))
    }
    sb.WriteString("\n")

    // all_configs/telegram
    sb.WriteString("## 📡 all_configs/telegram (کانفیگ‌های روزانه تلگرام)\n\n")
    sb.WriteString("| نام فایل | حجم | تعداد خطوط | تعداد کل پروکسی | آخرین تغییر | لینک خام (Raw) |\n")
    sb.WriteString("|----------|------|------------|----------------|--------------|----------------|\n")
    files, _ = filepath.Glob("📦 all_configs/📡 telegram/*.txt")
    for _, f := range files {
        name := filepath.Base(f)
        size, lines, mod := fileStats(f)
        validCount := countValidConfigs(f)
        url := fmt.Sprintf("%s/%s", baseRawURL, strings.ReplaceAll(f, "\\", "/"))
        sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%d` | `%d` | `%s` | [دانلود](%s) |\n",
            name, size, lines, validCount, mod.Format("2006-01-02 15:04:05"), url))
    }
    sb.WriteString("\n")

    // پوشه‌های اصلی
    for _, folder := range []string{"📡 telegram", "🔗 subscription", "🌍 mixed"} {
        sb.WriteString(fmt.Sprintf("## 🗂️ %s (خروجی اصلی)\n\n", folder))
        sb.WriteString("| نام فایل | حجم | تعداد خطوط | تعداد کل پروکسی | آخرین تغییر | لینک خام (Raw) |\n")
        sb.WriteString("|----------|------|------------|----------------|--------------|----------------|\n")
        files, _ = filepath.Glob(fmt.Sprintf("%s/*.txt", folder))
        for _, f := range files {
            name := filepath.Base(f)
            size, lines, mod := fileStats(f)
            validCount := countValidConfigs(f)
            url := fmt.Sprintf("%s/%s", baseRawURL, strings.ReplaceAll(f, "\\", "/"))
            sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%d` | `%d` | `%s` | [دانلود](%s) |\n",
                name, size, lines, validCount, mod.Format("2006-01-02 15:04:05"), url))
        }
        sb.WriteString("\n")
    }

    // آرشیو روزانه و لینک ZIP
    today := now.Format("2006-01-02")
    archiveZip := fmt.Sprintf("daily_archive_%s.zip", today)
    if _, err := os.Stat(archiveZip); err == nil {
        sb.WriteString("## 📦 آرشیو روزانه (فشرده)\n\n")
        sb.WriteString(fmt.Sprintf("| نام فایل | حجم | لینک دانلود |\n"))
        sb.WriteString(fmt.Sprintf("|----------|------|-------------|\n"))
        info, _ := os.Stat(archiveZip)
        sizeStr := formatBytes(info.Size())
        url := fmt.Sprintf("%s/%s", baseRawURL, archiveZip)
        sb.WriteString(fmt.Sprintf("| `%s` | `%s` | [دانلود ZIP](%s) |\n", archiveZip, sizeStr, url))
        sb.WriteString("\n")
    }

    // فضای ذخیره‌سازی
    totalSize := getTotalSize("📦 all_configs") + getTotalSize("🗄️ daily_archive")
    sb.WriteString("## 💾 فضای ذخیره‌سازی اشغال شده\n\n")
    sb.WriteString(fmt.Sprintf("| **کل فضای اشغالی** | `%s` |\n", formatBytes(totalSize)))
    sb.WriteString("\n---\n✅ این فایل به صورت خودکار هر بار تولید می‌شود.\n")

    os.MkdirAll("reports", 0755)
    os.WriteFile("reports/links.md", []byte(sb.String()), 0644)
}
