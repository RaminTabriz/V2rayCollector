package source

import (
    "context"
    "encoding/csv"
    "fmt"
    "io"
    "net/http"
    "os"
    "strings"
    "sync"
    "time"

    "github.com/PuerkitoBio/goquery"
    "github.com/RaminTabriz/V2rayCollector/internal/fetcher"
    "github.com/RaminTabriz/V2rayCollector/internal/parser"
    "github.com/projectdiscovery/gologger"
)

// Telegram مدیریت دریافت از کانال‌های تلگرام
type Telegram struct {
    channels    []string
    concurrency int
}

// NewTelegram سازنده جدید از فایل CSV
func NewTelegram(csvPath string, concurrency int) *Telegram {
    channels := readChannelsFromCSV(csvPath)
    return &Telegram{
        channels:    channels,
        concurrency: concurrency,
    }
}

// readChannelsFromCSV خواندن لیست کانال‌ها از فایل CSV (ستون اول URL)
func readChannelsFromCSV(path string) []string {
    file, err := os.Open(path)
    if err != nil {
        gologger.Warning().Msgf("Cannot open %s: %v", path, err)
        return nil
    }
    defer file.Close()
    reader := csv.NewReader(file)
    var channels []string
    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            continue
        }
        if len(record) > 0 && strings.HasPrefix(record[0], "https://t.me/") {
            channels = append(channels, record[0])
        }
    }
    return channels
}

// FetchAll دریافت همه کانال‌ها به صورت همزمان
func (t *Telegram) FetchAll(ctx context.Context, onConfig func(cfg, channel string)) {
    if len(t.channels) == 0 {
        gologger.Warning().Msg("No Telegram channels to fetch")
        return
    }
    gologger.Info().Msgf("Fetching %d Telegram channels", len(t.channels))

    jobs := make(chan string, len(t.channels))
    var wg sync.WaitGroup
    for i := 0; i < t.concurrency; i++ {
        wg.Add(1)
        go t.worker(ctx, jobs, &wg, onConfig)
    }
    for _, ch := range t.channels {
        select {
        case <-ctx.Done():
            close(jobs)
            return
        case jobs <- ch:
        }
    }
    close(jobs)
    wg.Wait()
}

// worker کارگر برای دریافت یک کانال
func (t *Telegram) worker(ctx context.Context, jobs <-chan string, wg *sync.WaitGroup, onConfig func(string, string)) {
    defer wg.Done()
    for url := range jobs {
        select {
        case <-ctx.Done():
            return
        default:
        }
        t.fetchOne(ctx, url, onConfig)
        time.Sleep(2 * time.Second) // فاصله بین درخواست‌ها
    }
}

// fetchOne دریافت و پردازش یک کانال تلگرام
func (t *Telegram) fetchOne(ctx context.Context, channelURL string, onConfig func(string, string)) {
    // محدودیت نرخ درخواست
    fetcher.WaitTelegram()

    channelName := extractChannelName(channelURL)
    if channelName == "" {
        gologger.Debug().Msgf("Invalid channel URL: %s", channelURL)
        return
    }

    webURL := fmt.Sprintf("https://t.me/s/%s", channelName)
    req, err := http.NewRequestWithContext(ctx, "GET", webURL, nil)
    if err != nil {
        gologger.Debug().Msgf("Failed to create request for %s: %v", webURL, err)
        return
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

    resp, err := fetcher.Client.Do(req)
    if err != nil {
        gologger.Debug().Msgf("Failed to fetch %s: %v", webURL, err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        gologger.Debug().Msgf("HTTP %d for %s", resp.StatusCode, webURL)
        return
    }

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        gologger.Debug().Msgf("Failed to parse HTML for %s: %v", webURL, err)
        return
    }

    var texts []string
    doc.Find(".tgme_widget_message_text, pre, code").Each(func(i int, s *goquery.Selection) {
        text := strings.TrimSpace(s.Text())
        if text != "" {
            texts = append(texts, text)
        }
    })

    collected := 0
    for _, text := range texts {
        configs := parser.ExtractAllConfigs(text)
        for _, cfg := range configs {
            onConfig(cfg, channelName)
            collected++
        }
    }
    if collected > 0 {
        gologger.Debug().Msgf("Collected %d configs from %s", collected, channelName)
    }
}

// extractChannelName استخراج نام کانال از URL
func extractChannelName(rawURL string) string {
    // قالب‌های ممکن: https://t.me/s/ChannelName , https://t.me/ChannelName
    if idx := strings.Index(rawURL, "/s/"); idx != -1 {
        part := rawURL[idx+3:]
        if slash := strings.Index(part, "/"); slash != -1 {
            return part[:slash]
        }
        return part
    }
    if idx := strings.Index(rawURL, "/t.me/"); idx != -1 {
        part := rawURL[idx+6:]
        if slash := strings.Index(part, "/"); slash != -1 {
            return part[:slash]
        }
        return part
    }
    return ""
}
