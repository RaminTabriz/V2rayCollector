package source

import (
    "compress/gzip"
    "context"
    "encoding/base64"
    "encoding/json"
    "io"
    "os"
    "strings"
    "sync"
    "time"

    "github.com/RaminTabriz/V2rayCollector/internal/fetcher"
    "github.com/RaminTabriz/V2rayCollector/internal/parser"
    "github.com/projectdiscovery/gologger"
)

// Subscription مدیریت دریافت از ساب‌لینک‌ها
type Subscription struct {
    urls        []string
    concurrency int
}

// NewSubscription سازنده جدید از فایل JSON
func NewSubscription(jsonPath string, concurrency int) *Subscription {
    urls := readSourcesFromJSON(jsonPath)
    return &Subscription{
        urls:        urls,
        concurrency: concurrency,
    }
}

// readSourcesFromJSON خواندن لیست ساب‌لینک‌ها از فایل JSON
func readSourcesFromJSON(path string) []string {
    data, err := os.ReadFile(path)
    if err != nil {
        gologger.Warning().Msgf("Cannot read %s: %v", path, err)
        return nil
    }
    var sources []string
    if err := json.Unmarshal(data, &sources); err != nil {
        gologger.Error().Msgf("Invalid JSON in %s: %v", path, err)
        return nil
    }
    return sources
}

// FetchAll دریافت تمام ساب‌لینک‌ها به صورت همزمان
func (s *Subscription) FetchAll(ctx context.Context, onConfig func(cfg string)) {
    if len(s.urls) == 0 {
        gologger.Warning().Msg("No subscription sources to fetch")
        return
    }
    gologger.Info().Msgf("Fetching %d subscription sources", len(s.urls))

    jobs := make(chan string, len(s.urls))
    var wg sync.WaitGroup
    for i := 0; i < s.concurrency; i++ {
        wg.Add(1)
        go s.worker(ctx, jobs, &wg, onConfig)
    }
    for _, u := range s.urls {
        select {
        case <-ctx.Done():
            close(jobs)
            return
        case jobs <- u:
        }
    }
    close(jobs)
    wg.Wait()
}

// worker کارگر برای دریافت یک ساب‌لینک
func (s *Subscription) worker(ctx context.Context, jobs <-chan string, wg *sync.WaitGroup, onConfig func(string)) {
    defer wg.Done()
    for url := range jobs {
        select {
        case <-ctx.Done():
            return
        default:
        }
        s.fetchOne(ctx, url, onConfig)
        time.Sleep(1 * time.Second) // احترام به منابع
    }
}

// fetchOne دریافت و پردازش یک ساب‌لینک
func (s *Subscription) fetchOne(ctx context.Context, url string, onConfig func(string)) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        gologger.Debug().Msgf("Failed to create request for %s: %v", url, err)
        return
    }
    req.Header.Set("Accept-Encoding", "gzip")

    resp, err := fetcher.Client.Do(req)
    if err != nil {
        gologger.Debug().Msgf("Failed to fetch subscription %s: %v", url, err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        gologger.Debug().Msgf("Subscription %s returned HTTP %d", url, resp.StatusCode)
        return
    }

    var reader io.Reader = resp.Body
    if resp.Header.Get("Content-Encoding") == "gzip" {
        gr, err := gzip.NewReader(resp.Body)
        if err == nil {
            defer gr.Close()
            reader = gr
        }
    }

    body, err := io.ReadAll(reader)
    if err != nil {
        gologger.Debug().Msgf("Failed to read body of %s: %v", url, err)
        return
    }

    content := string(body)

    // تلاش برای دیکد base64 (برخی ساب‌لینک‌ها base64 هستند)
    trimmed := strings.TrimSpace(content)
    if decoded, err := base64.StdEncoding.DecodeString(trimmed); err == nil && len(decoded) > 0 {
        content = string(decoded)
    }

    configs := parser.ExtractAllConfigs(content)
    gologger.Debug().Msgf("Extracted %d configs from %s", len(configs), url)

    for _, cfg := range configs {
        onConfig(cfg)
    }
}
