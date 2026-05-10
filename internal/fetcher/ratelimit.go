package fetcher

import (
    "time"
)

var telegramLimiter = time.NewTicker(200 * time.Millisecond) // 5 requests per second

func WaitTelegram() {
    <-telegramLimiter.C
}
