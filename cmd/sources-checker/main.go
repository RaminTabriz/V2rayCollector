package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "math/rand"
    "net/http"
    "os"
    "sort"
    "strings"
    "sync"
    "time"
)

const (
    sampleSize          = 50 * 1024
    defaultRetryCount   = 3
    defaultBaseDelay    = 1 * time.Second
    defaultJitter       = 500 * time.Millisecond
    defaultActiveDays   = 30
    defaultConcurrency  = 5
    dataDir             = "data"
    deadSourcesRecent   = dataDir + "/dead_sources_recent.json"
    deadSourcesOld      = dataDir + "/dead_sources_old.json"
)

var (
    oldScan = flag.Bool("old-scan", false, "Scan only sources older than 365 days (yearly)")
    client  = &http.Client{Timeout: 10 * time.Second}
)

type DeadSourceInfo struct {
    URL       string `json:"url"`
    LastMod   int64  `json:"last_mod"`
    CheckedAt int64  `json:"checked_at"`
}

type SourceStatus struct {
    URL       string    `json:"url"`
    LastMod   time.Time `json:"last_mod"`
    HasConfig bool      `json:"has_config"`
    Status    string    `json:"status"`
    Error     string    `json:"error,omitempty"`
}

func main() {
    flag.Parse()
    os.MkdirAll(dataDir, 0755)
    os.MkdirAll("reports", 0755)

    if len(os.Args) < 2 {
        fmt.Println("Usage: go run sources-checker.go <Sources.json> [-old-scan]")
        os.Exit(1)
    }
    inputFile := os.Args[1]

    data, err := os.ReadFile(inputFile)
    if err != nil {
        fmt.Printf("Error reading %s: %v\n", inputFile, err)
        os.Exit(1)
    }
    var sources []string
    if err := json.Unmarshal(data, &sources); err != nil {
        fmt.Printf("Error parsing JSON: %v\n", err)
        os.Exit(1)
    }
    if len(sources) == 0 {
        fmt.Println("No sources found.")
        return
    }

    recentDead := loadDeadSources(deadSourcesRecent)
    oldDead := loadDeadSources(deadSourcesOld)

    urlSet := make(map[string]bool)
    for _, src := range sources {
        urlSet[src] = true
    }
    if *oldScan {
        for url := range oldDead {
            urlSet[url] = true
        }
    } else {
        for url := range recentDead {
            urlSet[url] = true
        }
    }
    urlList := make([]string, 0, len(urlSet))
    for u := range urlSet {
        urlList = append(urlList, u)
    }

    jobs := make(chan string, len(urlList))
    results := make(chan SourceStatus, len(urlList))
    var wg sync.WaitGroup
    for i := 0; i < defaultConcurrency; i++ {
        wg.Add(1)
        go worker(jobs, results, &wg)
    }
    for _, url := range urlList {
        jobs <- url
    }
    close(jobs)
    wg.Wait()
    close(results)

    var activeURLs []string
    updatedRecent := make(map[string]DeadSourceInfo)
    updatedOld := make(map[string]DeadSourceInfo)

    for res := range results {
        daysSince := 0
        if !res.LastMod.IsZero() {
            daysSince = int(time.Since(res.LastMod).Hours() / 24)
        }
        if res.Status == "OK" && res.HasConfig && daysSince <= defaultActiveDays {
            activeURLs = append(activeURLs, res.URL)
            delete(recentDead, res.URL)
            delete(oldDead, res.URL)
        } else {
            info := DeadSourceInfo{
                URL:       res.URL,
                LastMod:   res.LastMod.Unix(),
                CheckedAt: time.Now().Unix(),
            }
            if daysSince > 365 {
                updatedOld[res.URL] = info
                delete(recentDead, res.URL)
            } else {
                updatedRecent[res.URL] = info
                delete(oldDead, res.URL)
            }
        }
    }
    for k, v := range recentDead {
        updatedRecent[k] = v
    }
    for k, v := range oldDead {
        updatedOld[k] = v
    }

    saveActiveSources(activeURLs, inputFile)
    saveDeadSources(deadSourcesRecent, updatedRecent)
    saveDeadSources(deadSourcesOld, updatedOld)
    fmt.Printf("\n✅ Active sources: %d, Recent dead: %d, Old dead: %d\n", len(activeURLs), len(updatedRecent), len(updatedOld))
}

func worker(jobs <-chan string, results chan<- SourceStatus, wg *sync.WaitGroup) {
    defer wg.Done()
    for url := range jobs {
        res := checkSourceWithRetry(url)
        results <- res
    }
}

func checkSourceWithRetry(url string) SourceStatus {
    var lastErr error
    for attempt := 1; attempt <= defaultRetryCount; attempt++ {
        res, err := checkSource(url)
        if err == nil {
            return res
        }
        lastErr = err
        delay := defaultBaseDelay * time.Duration(1<<uint(attempt-1))
        jitter := time.Duration(rand.Int63n(int64(defaultJitter)))
        time.Sleep(delay + jitter)
    }
    return SourceStatus{URL: url, Status: "DEAD", Error: lastErr.Error()}
}

func checkSource(url string) (SourceStatus, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return SourceStatus{}, err
    }
    req.Header.Set("User-Agent", "Mozilla/5.0")
    req.Header.Set("Range", "bytes=0-50000")
    resp, err := client.Do(req)
    if err != nil {
        return SourceStatus{}, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 && resp.StatusCode != 206 {
        return SourceStatus{URL: url, Status: "DEAD"}, nil
    }
    lastModStr := resp.Header.Get("Last-Modified")
    var lastMod time.Time
    if lastModStr != "" {
        lastMod, _ = time.Parse(time.RFC1123, lastModStr)
    }
    limited := io.LimitReader(resp.Body, sampleSize)
    body, err := io.ReadAll(limited)
    if err != nil {
        return SourceStatus{}, err
    }
    content := string(body)
    hasConfig := anyConfigInText(content)
    daysSince := 0
    if !lastMod.IsZero() {
        daysSince = int(time.Since(lastMod).Hours() / 24)
    }
    status := "DEAD"
    if resp.StatusCode == 200 || resp.StatusCode == 206 {
        if hasConfig && daysSince <= defaultActiveDays {
            status = "OK"
        } else if !hasConfig {
            status = "NO_CONFIG"
        } else {
            status = "INACTIVE"
        }
    }
    fmt.Printf("[INFO] %s -> lastMod: %s (%d days), hasConfig: %v, status: %s\n",
        url, lastMod.Format("2006-01-02"), daysSince, hasConfig, status)
    return SourceStatus{
        URL:       url,
        LastMod:   lastMod,
        HasConfig: hasConfig,
        Status:    status,
    }, nil
}

func anyConfigInText(text string) bool {
    patterns := []string{
        `vmess://`, `vless://`, `trojan://`, `ss://`, `ssr://`,
        `hysteria2://`, `hy2://`, `tuic://`, `wireguard://`,
        `tg://proxy`, `tg://socks`, `slipnet://`,
        `socks://`, `socks5://`, `http://`, `https://`,
        `-----BEGIN ARGO VPN BRIDGE BLOCK-----`,
    }
    for _, p := range patterns {
        if strings.Contains(text, p) {
            return true
        }
    }
    return false
}

func loadDeadSources(file string) map[string]DeadSourceInfo {
    m := make(map[string]DeadSourceInfo)
    data, err := os.ReadFile(file)
    if err != nil {
        return m
    }
    var list []DeadSourceInfo
    if err := json.Unmarshal(data, &list); err != nil {
        return m
    }
    for _, item := range list {
        m[item.URL] = item
    }
    return m
}

func saveDeadSources(file string, m map[string]DeadSourceInfo) {
    list := make([]DeadSourceInfo, 0, len(m))
    for _, v := range m {
        list = append(list, v)
    }
    sort.Slice(list, func(i, j int) bool { return list[i].URL < list[j].URL })
    data, _ := json.MarshalIndent(list, "", "  ")
    os.WriteFile(file, data, 0644)
    fmt.Printf("✅ Saved %s with %d entries.\n", file, len(list))
}

func saveActiveSources(active []string, inputFile string) {
    data, err := json.MarshalIndent(active, "", "  ")
    if err != nil {
        fmt.Printf("Error marshalling JSON: %v\n", err)
        return
    }
    os.WriteFile(inputFile, data, 0644)
    fmt.Printf("✅ Active sources written to %s\n", inputFile)
}
