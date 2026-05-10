<div dir="rtl" align="center">

# 🚀 V2rayCollector – جمع‌آوری خودکار کانفیگ V2Ray/Xray از تلگرام، ساب‌لینک و فورک‌های گیت‌هاب

[![GitHub release](https://img.shields.io/github/v/release/RaminTabriz/V2rayCollector?style=flat-square&logo=github)](https://github.com/RaminTabriz/V2rayCollector/releases)
[![GitHub Repo stars](https://img.shields.io/github/stars/RaminTabriz/V2rayCollector?style=flat-square&logo=github)](https://github.com/RaminTabriz/V2rayCollector/stargazers)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/RaminTabriz/V2rayCollector/Collector.yml?branch=main&style=flat-square&logo=githubactions)](https://github.com/RaminTabriz/V2rayCollector/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/RaminTabriz/V2rayCollector?style=flat-square)](https://goreportcard.com/report/github.com/RaminTabriz/V2rayCollector)
[![License](https://img.shields.io/github/license/RaminTabriz/V2rayCollector?style=flat-square)](LICENSE)

**یک ربات تمام‌خودکار برای جمع‌آوری کانفیگ‌های رایگان پروکسی**  
⚡ بدون نیاز به سرور شخصی – فقط با گیت‌هاب اکشنز ⚡

</div>

---

## 🧠 درباره پروژه

پروژه **V2rayCollector** با زبان Go نوشته شده و از GitHub Actions برای اجرای دوره‌ای استفاده می‌کند. وظیفه آن:

- اسکرپ کردن کانال‌های تلگرام (لیست شده در `channels.csv`)
- دانلود ساب‌لینک‌ها (لیست شده در `Sources.json`)
- اسکن فورک‌های یک مخزن هدف گیت‌هاب برای یافتن فایل‌های ساب‌اسکریپشن

سپس کانفیگ‌های استخراج شده را **پردازش، ددابلیکیت (بر اساس اثر انگشت)، فیلتر امنیتی** (حذف `allowInsecure=true` و `encryption=none`) و در نهایت **دسته‌بندی بر اساس پروتکل** و ذخیره در پوشه‌های خروجی (با ایموجی) می‌کند.

---

## ✨ قابلیت‌ها

- **پشتیبانی از پروتکل‌های متعدد**  
  📦 Vmess – 🕳️ VLess – 🐴 Trojan – 🐍 Shadowsocks (ss) – 🔄 ShadowsocksR (ssr) – ⚡ Hysteria2 – 🧩 Tuic – 🔒 WireGuard – 🌌 WARP – 📱 MTProto Proxy – 🧦 SOCKS5 – 🧦 SOCKS – 🌐 HTTP – 🔒 HTTPS – ☁️ Argo – 🕸️ Slipnet – 🛡️ Invizible – 🎭 Mixed

- **دریافت از سه منبع اصلی**  
  - کانال‌های تلگرام (با RSS و HTML fallback)  
  - ساب‌لینک‌های متن (JSON, TXT, base64, gzip)  
  - فورک‌های مخزن هدف در گیت‌هاب

- **پردازش هوشمند**  
  - تشخیص پروتکل با regex  
  - فیلتر امنیتی (حذف کانفیگ‌های ناامن)  
  - ددابلیکیت پیشرفته با اثر انگشت (برای VMess از فیلدهای JSON، برای سایر از host+port+user)  
  - جداسازی کانفیگ‌های چسبیده (درج newline قبل از هر پروتکل)

- **خروجی‌های منظم**  
  - دسته‌بندی بر اساس پروتکل و منبع  
  - استفاده از ایموجی برای نام پوشه‌ها و فایل‌ها  
  - آرشیو روزانه (`🗄️ daily_archive/`) برای حفظ تاریخچه

- **خودکارسازی کامل با GitHub Actions**  
  - اجرای اصلی هر 20 دقیقه  
  - اسکن روزانه کانال‌ها و جداسازی کانال‌های مرده  
  - احیای خودکار کانال‌های دوباره فعال شده  
  - بهینه‌سازی هفتگی ساب‌لینک‌ها

---

## 📁 ساختار پروژه (پس از بازطراحی)

```
V2rayCollector/
├── cmd/
│   ├── collector/                 # جمع‌آوری‌کننده اصلی
│   ├── channel-scanner/           # اسکنر کانال‌های تلگرام (جدا کردن مرده‌ها)
│   ├── sources-checker/           # اسکنر ساب‌لینک‌ها
│   └── revive-scanner/            # احیای کانال‌های مرده
├── internal/
│   ├── cache/                     # مدیریت کش و ددابلیکیت
│   ├── parser/                    # تشخیص پروتکل، فیلتر امنیتی، استخراج کانفیگ
│   ├── fetcher/                   # درخواست‌های HTTP با proxy و rate limit
│   ├── source/                    # دریافت از تلگرام، ساب‌لینک، فورک گیت‌هاب
│   ├── output/                    # نوشتن فایل‌های خروجی و آرشیو
│   └── report/                    # تولید گزارش‌های آماری و لینک‌ها
├── .github/workflows/             # فایل‌های GitHub Actions
├── 📡 telegram/                   (تولید می‌شود)
├── 🔗 subscription/               (تولید می‌شود)
├── 🌍 mixed/                      (تولید می‌شود)
├── 📦 all_configs/                (تولید می‌شود)
├── 🗄️ daily_archive/              (تولید می‌شود)
├── reports/                       (تولید می‌شود)
├── data/                          (تولید می‌شود)
├── channels.csv                   (ورودی)
├── Sources.json                   (ورودی)
├── config_cache.json              (تولید می‌شود)
└── README.md
```

---

## 🚀 نصب و راه‌اندازی (برای استفاده شخصی)

### 1. کلون مخزن

```bash
git clone https://github.com/RaminTabriz/V2rayCollector.git
cd V2rayCollector
```

### 2. نصب وابستگی‌ها

```bash
go mod download
go mod tidy
```

### 3. آماده‌سازی فایل‌های ورودی

- **`channels.csv`** – لیست کانال‌های تلگرام (یک ستون `URL`، ستون دوم دلخواه)  
  مثال:  
  ```
  URL,AllMessagesFlag
  https://t.me/s/FreeV2rays,false
  https://t.me/s/ConfigX2ray,false
  ```
- **`Sources.json`** – لیست ساب‌لینک‌ها (آرایه‌ای از آدرس‌ها)  
  مثال:  
  ```json
  [
    "https://raw.githubusercontent.com/Epodonios/v2ray-configs/main/All_Configs_Sub.txt",
    "https://raw.githubusercontent.com/wang1680/v2ray-configs/main/all_configs.txt"
  ]
  ```

### 4. اجرای دستی (تست)

```bash
# اجرای اصلی (جمع‌آوری از سه منبع)
go run ./cmd/collector -channels channels.csv -sources Sources.json -sort -clash

# اسکن کانال‌ها و به‌روزرسانی channels.csv
go run ./cmd/channel-scanner -input channels.csv -output channels.csv

# احیای کانال‌های مرده
go run ./cmd/revive-scanner

# بهینه‌سازی ساب‌لینک‌ها
go run ./cmd/sources-checker Sources.json
```

### 5. تنظیم GitHub Actions (اجرای خودکار)

فایل‌های workflow در پوشه `.github/workflows/` آماده هستند. پس از push به مخزن، هر workflow در زمان مشخص شده اجرا می‌شود. برای فعالسازی کامل:

- مخزن خود را در گیت‌هاب ایجاد کنید.
- کدها را push کنید.
- در تنظیمات مخزن (`Settings > Actions > General`) اجازه `Read and write permissions` را بدهید.
- (اختیاری) برای دریافت گزارش در تلگرام، secrets به نام‌های `TELEGRAM_BOT_TOKEN` و `TELEGRAM_CHAT_ID` را اضافه کنید.

---

## ⚙️ راهنمای فایل‌های پیکربندی

| فایل | توضیح |
|------|-------|
| `channels.csv` | لیست کانال‌های تلگرام. فقط کانال‌های فعال (با پست جدید و حاوی کانفیگ) در این فایل باقی می‌مانند. کانال‌های مرده به `data/dead_channels_*.json` منتقل می‌شوند. |
| `Sources.json` | لیست ساب‌لینک‌ها (لینک مستقیم به فایل‌های متنی حاوی کانفیگ). منابع مرده به `data/dead_sources_*.json` منتقل می‌شوند. |
| `config_cache.json` | کش کانفیگ‌ها با اثر انگشت. به صورت خودکار مدیریت می‌شود. |
| `last_archive_time.txt` | زمان آخرین آرشیو روزانه. برای جلوگیری از دوباره‌نویسی استفاده می‌شود. |

---

## 🛠️ فلگ‌های خط فرمان (برای `cmd/collector`)

| فلگ | پیش‌فرض | توضیح |
|-----|---------|-------|
| `-channels` | `channels.csv` | مسیر فایل CSV کانال‌ها |
| `-sources` | `Sources.json` | مسیر فایل JSON ساب‌لینک‌ها |
| `-concurrent` | `3` | تعداد workerهای همزمان |
| `-fork-scan` | `true` | اسکن فورک‌های گیت‌هاب |
| `-target-repo` | `mahsanet/MahsaFreeConfig` | مخزن هدف برای اسکن فورک |
| `-sort` | `false` | مرتب‌سازی کانفیگ‌ها بر اساس زمان |
| `-clash` | `false` | تولید فایل `clash-config.yaml` |

---

## 📊 خروجی‌ها

| پوشه/فایل | توضیح |
|-----------|-------|
| `📡 telegram/` | کانفیگ‌های استخراج شده از تلگرام، هر کانال در زیرپوشه خود و تفکیک پروتکل |
| `🔗 subscription/` | کانفیگ‌های ساب‌لینک بر اساس پروتکل |
| `🌍 mixed/` | کانفیگ‌هایی که پروتکل آنها تشخیص داده نشده |
| `📦 all_configs/` | انباشت روزانه کانفیگ‌های جدید (از زمان آخرین آرشیو) |
| `🗄️ daily_archive/` | آرشیو کامل روزهای قبل |
| `reports/collector_stats.md` | گزارش آماری (تعداد کانفیگ، تفکیک پروتکل، کانال‌های پربازده) |
| `reports/links.md` | لینک‌های دانلود مستقیم تمام فایل‌های خروجی |
| `clash-config.yaml` | فایل پیکربندی Clash (در صورت استفاده از فلگ `-clash`) |

---

## 🤝 مشارکت

هرگونه issue یا pull request با استقبال مواجه می‌شود. لطفاً قبل از تغییرات عمده، یک issue باز کنید تا درباره آن گفتگو کنیم.

---

## 📜 مجوز

این پروژه تحت مجوز **MIT** منتشر شده است. برای جزئیات به فایل `LICENSE` مراجعه کنید.

---

## 💬 تماس و پشتیبانی

- کانال تلگرام: [ID](UR) (در صورت وجود)
- ایمیل: XXX@XXX.com (نمادین)

---

<div dir="rtl" align="center">

**اگر این پروژه برایتان مفید بود، با ⭐ ستاره دادن به ما انگیزه بدهید!**  
**از همراهی شما سپاسگزاریم.**

</div>
```

---
