package source

import (
    "context"
    "encoding/csv"
    "fmt"
    "io"
    "os"
    "strings"
    "sync"
    "time"

    "github.com/PuerkitoBio/goquery"
    "github.com/ramin00542/GO_V2rayCollector/internal/fetcher"
    "github.com/ramin00542/GO_V2rayCollector/internal/parser"
    "github.com/projectdiscovery/gologger"
)

type Telegram struct {
    channels    []string
    concurrency int
}

func NewTelegram(csvPath string, concurrency int) *Telegram {
    channels := readChannelsFromCSV(csvPath)
    return &Telegram{
        channels:    channels,
        concurrency: concurrency,
    }
}

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

func (t *Telegram) worker(ctx context.Context, jobs <-chan string, wg *sync.WaitGroup, onConfig func(string, string)) {
    defer wg.Done()
    for url := range jobs {
        select {
        case <-ctx.Done():
            return
        default:
        }
        fetcher.WaitTelegram()
        channelName := extractChannelName(url)
        if channelName == "" {
            continue
        }
        webURL := fmt.Sprintf("https://t.me/s/%s", channelName)
        resp, err := fetcher.Client.Get(webURL)
        if err != nil {
            gologger.Debug().Msgf("Failed to fetch %s: %v", webURL, err)
            continue
        }
        if resp.StatusCode != 200 {
            resp.Body.Close()
            continue
        }
        doc, err := goquery.NewDocumentFromReader(resp.Body)
        resp.Body.Close()
        if err != nil {
            continue
        }
        var texts []string
        doc.Find(".tgme_widget_message_text, pre, code").Each(func(i int, s *goquery.Selection) {
            texts = append(texts, s.Text())
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
        time.Sleep(2 * time.Second)
    }
}

func extractChannelName(rawURL string) string {
    // formats: https://t.me/s/ChannelName or https://t.me/ChannelName
    if idx := strings.Index(rawURL, "/s/"); idx != -1 {
        parts := strings.Split(rawURL[idx+3:], "/")
        if len(parts) > 0 {
            return parts[0]
        }
    }
    if idx := strings.Index(rawURL, "/t.me/"); idx != -1 {
        after := rawURL[idx+6:]
        if slash := strings.Index(after, "/"); slash != -1 {
            return after[:slash]
        }
        return after
    }
    return ""
}
