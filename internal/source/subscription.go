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

    "github.com/ramin00542/GO_V2rayCollector/internal/fetcher"
    "github.com/ramin00542/GO_V2rayCollector/internal/parser"
    "github.com/projectdiscovery/gologger"
)

type Subscription struct {
    urls        []string
    concurrency int
}

func NewSubscription(jsonPath string, concurrency int) *Subscription {
    urls := readSourcesFromJSON(jsonPath)
    return &Subscription{
        urls:        urls,
        concurrency: concurrency,
    }
}

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

func (s *Subscription) worker(ctx context.Context, jobs <-chan string, wg *sync.WaitGroup, onConfig func(string)) {
    defer wg.Done()
    for url := range jobs {
        select {
        case <-ctx.Done():
            return
        default:
        }
        resp, err := fetcher.Client.Get(url)
        if err != nil {
            gologger.Debug().Msgf("Failed to fetch subscription %s: %v", url, err)
            continue
        }
        if resp.StatusCode != 200 {
            resp.Body.Close()
            continue
        }
        var reader io.Reader = resp.Body
        if resp.Header.Get("Content-Encoding") == "gzip" {
            gr, err := gzip.NewReader(resp.Body)
            if err == nil {
                reader = gr
                defer gr.Close()
            }
        }
        body, err := io.ReadAll(reader)
        resp.Body.Close()
        if err != nil {
            continue
        }
        content := string(body)
        // Try base64 decode if needed
        if decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(content)); err == nil && len(decoded) > 0 {
            content = string(decoded)
        }
        configs := parser.ExtractAllConfigs(content)
        gologger.Debug().Msgf("Extracted %d configs from %s", len(configs), url)
        for _, cfg := range configs {
            onConfig(cfg)
        }
        time.Sleep(1 * time.Second)
    }
}
