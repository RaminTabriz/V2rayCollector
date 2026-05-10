package fetcher

import (
    "sync"
    "time"
)

var (
    // telegramLimiter برای محدود کردن نرخ درخواست به تلگرام (5 درخواست در ثانیه)
    telegramLimiter *time.Ticker
    // mu برای جلوگیری از race condition در مقداردهی
    mu sync.Mutex
)

// InitRateLimiter مقداردهی اولیه محدودکننده نرخ درخواست
// ratePerSecond: تعداد درخواست مجاز در ثانیه (پیش‌فرض 5)
func InitRateLimiter(ratePerSecond int) {
    mu.Lock()
    defer mu.Unlock()
    if telegramLimiter != nil {
        telegramLimiter.Stop()
    }
    interval := time.Second / time.Duration(ratePerSecond)
    telegramLimiter = time.NewTicker(interval)
}

// WaitTelegram منتظر می‌ماند تا نوبت درخواست بعدی به تلگرام برسد
// اگر محدودکننده مقداردهی نشده باشد، یک نمونه پیش‌فرض (5 تیک در ثانیه) ایجاد می‌کند
func WaitTelegram() {
    mu.Lock()
    if telegramLimiter == nil {
        // مقداردهی پیش‌فرض: 5 درخواست در ثانیه
        telegramLimiter = time.NewTicker(time.Second / 5)
    }
    mu.Unlock()
    <-telegramLimiter.C
}

// StopRateLimiter توقف تیکر (اختیاری، معمولاً نیازی نیست)
func StopRateLimiter() {
    mu.Lock()
    defer mu.Unlock()
    if telegramLimiter != nil {
        telegramLimiter.Stop()
        telegramLimiter = nil
    }
}
