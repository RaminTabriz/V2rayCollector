package main

import (
    "encoding/csv"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "math/rand"
    "net/http"
    "os"
    "regexp"
    "sort"
    "strings"
    "sync"
    "time"

    "github.com/PuerkitoBio/goquery"
)

const (
    defaultActiveDays   = 30
    defaultConcurrency  = 5
    defaultRetryCount   = 3
    defaultBaseDelay    = 1 * time.Second
    defaultJitter       = 500 * time.Millisecond
    dataDir             = "data"
    deadChannelsRecent  = dataDir + "/dead_channels_recent.json"
    deadChannelsOld     = dataDir + "/dead_channels_old.json"
)

var (
    inputCSV      = flag.String("input", "channels.csv", "Input CSV file")
    outputCSV     = flag.String("output", "channels.csv", "Output CSV file")
    activeDays    = flag.Int("active-days", defaultActiveDays, "Max inactive days")
    concurrency   = flag.Int("concurrency", defaultConcurrency, "Number of concurrent workers")
    oldScan       = flag.Bool("old-scan", false, "Scan only channels older than 365 days (yearly)")

    client = &http.Client{Timeout: 15 * time.Second}

    // الگوهای تشخیص کانفیگ (پشتیبانی از تمام پروتکل‌ها)
    configPatterns = []*regexp.Regexp{
        regexp.MustCompile(`vmess://[A-Za-z0-9+/]+={0,2}(?:\?[^\s]*)?`),
        regexp.MustCompile(`vless://[^\s]+`),
        regexp.MustCompile(`trojan://[^@\s]+@[^\s]+`),
        regexp.MustCompile(`ss://[A-Za-z0-9+/]+={0,2}@[^\s]+`),
        regexp.MustCompile(`ssr://[A-Za-z0-9+/=]+`),
        regexp.MustCompile(`hysteria2://[^\s]+`),
        regexp.MustCompile(`hy2://[^\s]+`),
        regexp.MustCompile(`tuic://[^\s]+`),
        regexp.MustCompile(`wireguard://[^\s]+`),
        regexp.MustCompile(`warp://[^\s]+`),
        regexp.MustCompile(`slipnet://[^\s]+`),
        regexp.MustCompile(`tg://proxy\?[^\s]+`),
        regexp.MustCompile(`tg://socks\?[^\s]+`),
        regexp.MustCompile(`https?://[^\s]+:\d+(?:[^\s]*)?`),
        regexp.MustCompile(`socks(?:5)?://[^\s]+@[^\s]+`),
        regexp.MustCompile(`socks(?:5)?://[^\s]+:\d+`),
        regexp.MustCompile(`-----BEGIN ARGO VPN BRIDGE BLOCK-----[\s\S]+?-----END ARGO VPN BRIDGE BLOCK-----`),
    }

    printMutex sync.Mutex
)

type DeadChannelInfo struct {
    URL       string `json:"url"`
    LastPost  int64  `json:"last_post"`
    CheckedAt int64  `json:"checked_at"`
}

type ScanResult struct {
    URL          string    `json:"url"`
    LastPost     time.Time `json:"last_post"`
    HasConfig    bool      `json:"has_config"`
    Status       string    `json:"status"`
    MessageCount int       `json:"msg_count"`
    Error        string    `json:"error,omitempty"`
    Timestamp    time.Time `json:"timestamp"`
}

func main() {
    flag.Parse()
    os.MkdirAll(dataDir, 0755)
    os.MkdirAll("reports", 0755)

    recentDead := loadDeadArchive(deadChannelsRecent)
    oldDead := loadDeadArchive(deadChannelsOld)

    records, headers, err := readCSV(*inputCSV)
    if err != nil {
        fmt.Printf("Error reading CSV: %v\n", err)
        os.Exit(1)
    }
    if len(records) == 0 {
        fmt.Println("No channels found.")
        return
    }

    // ساخت لیست URLها بر اساس flag old-scan
    urlSet := make(map[string]bool)
    for _, row := range records {
        if len(row) > 0 {
            urlSet[row[0]] = true
        }
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

    if len(urlList) == 0 {
        fmt.Println("No channels to scan.")
        return
    }

    jobs := make(chan string, len(urlList))
    results := make(chan ScanResult, len(urlList))
    var wg sync.WaitGroup
    for i := 0; i < *concurrency; i++ {
        wg.Add(1)
        go worker(jobs, results, &wg)
    }
    for _, url := range urlList {
        jobs <- url
    }
    close(jobs)
    wg.Wait()
    close(results)

    var activeList []ScanResult
    updatedRecent := make(map[string]DeadChannelInfo)
    updatedOld := make(map[string]DeadChannelInfo)

    for res := range results {
        if res.Status == "active" {
            activeList = append(activeList, res)
            delete(recentDead, res.URL)
            delete(oldDead, res.URL)
        } else {
            daysSince := 0
            if !res.LastPost.IsZero() {
                daysSince = int(time.Since(res.LastPost).Hours() / 24)
            }
            info := DeadChannelInfo{
                URL:       res.URL,
                LastPost:  res.LastPost.Unix(),
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

    // حفظ مواردی که اسکن نشده‌اند
    for k, v := range recentDead {
        updatedRecent[k] = v
    }
    for k, v := range oldDead {
        updatedOld[k] = v
    }

    // نوشتن CSV فقط با کانال‌های فعال
    if err := writeActiveCSV(*outputCSV, headers, activeList); err != nil {
        fmt.Printf("Error writing CSV: %v\n", err)
    } else {
        fmt.Printf("✅ Updated %s with %d active channels.\n", *outputCSV, len(activeList))
    }

    saveDeadArchive(deadChannelsRecent, updatedRecent)
    saveDeadArchive(deadChannelsOld, updatedOld)

    fmt.Printf("\n✅ Active: %d, Recent dead: %d, Old dead: %d\n", len(activeList), len(updatedRecent), len(updatedOld))
}

func worker(jobs <-chan string, results chan<- ScanResult, wg *sync.WaitGroup) {
    defer wg.Done()
    for url := range jobs {
        res := analyzeWithRetry(url)
        results <- res
    }
}

func analyzeWithRetry(url string) ScanResult {
    var lastErr error
    for attempt := 1; attempt <= defaultRetryCount; attempt++ {
        res, err := analyzeFull(url)
        if err == nil {
            return res
        }
        lastErr = err
        delay := defaultBaseDelay * time.Duration(1<<uint(attempt-1))
        jitter := time.Duration(rand.Int63n(int64(defaultJitter)))
        time.Sleep(delay + jitter)
    }
    return ScanResult{URL: url, Status: "error", Error: lastErr.Error(), Timestamp: time.Now()}
}

func analyzeFull(channelURL string) (ScanResult, error) {
    channelName := extractChannelName(channelURL)
    if channelName == "" {
        return ScanResult{}, fmt.Errorf("invalid URL")
    }

    // اول تلاش RSS
    rssURL := fmt.Sprintf("https://t.me/s/%s.rss", channelName)
    res, err := fetchFromRSS(rssURL, channelURL)
    if err == nil {
        return res, nil
    }
    // اگر RSS کار نکرد، HTML
    htmlURL := fmt.Sprintf("https://t.me/s/%s", channelName)
    return fetchFromHTML(htmlURL, channelURL)
}

func fetchFromRSS(rssURL, origURL string) (ScanResult, error) {
    resp, err := client.Get(rssURL)
    if err != nil {
        return ScanResult{}, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return ScanResult{}, fmt.Errorf("HTTP %d", resp.StatusCode)
    }
    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return ScanResult{}, err
    }
    var latestTime time.Time
    var anyConfig bool
    msgCount := doc.Find("item").Length()
    doc.Find("item").Each(func(i int, s *goquery.Selection) {
        pubDate := s.Find("pubDate").Text()
        if pubDate != "" {
            t, err := time.Parse(time.RFC1123Z, pubDate)
            if err == nil && (latestTime.IsZero() || t.After(latestTime)) {
                latestTime = t
            }
        }
        desc := s.Find("description").Text()
        if anyConfigInText(desc) {
            anyConfig = true
        }
    })
    if latestTime.IsZero() {
        return ScanResult{}, fmt.Errorf("no valid pubDate")
    }
    daysSince := int(time.Since(latestTime).Hours() / 24)
    status := "inactive"
    if anyConfig && daysSince <= *activeDays {
        status = "active"
    }
    safePrintf("[INFO] %s -> last: %s (%d days), config: %v, status: %s\n",
        origURL, latestTime.Format("2006-01-02"), daysSince, anyConfig, status)
    return ScanResult{
        URL:          origURL,
        LastPost:     latestTime,
        HasConfig:    anyConfig,
        Status:       status,
        MessageCount: msgCount,
        Timestamp:    time.Now(),
    }, nil
}

func fetchFromHTML(htmlURL, origURL string) (ScanResult, error) {
    resp, err := client.Get(htmlURL)
    if err != nil {
        return ScanResult{}, err
    }
    defer resp.Body.Close()
    if resp.StatusCode == 404 {
        return ScanResult{URL: origURL, Status: "banned", Error: "channel not found", Timestamp: time.Now()}, nil
    }
    if resp.StatusCode != 200 {
        return ScanResult{}, fmt.Errorf("HTTP %d", resp.StatusCode)
    }
    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return ScanResult{}, err
    }
    var lastTime time.Time
    doc.Find("time").Each(func(i int, s *goquery.Selection) {
        if i == 0 {
            if dt, ok := s.Attr("datetime"); ok {
                if t, err := time.Parse(time.RFC3339, dt); err == nil {
                    lastTime = t
                }
            }
        }
    })
    if lastTime.IsZero() {
        doc.Find(".datetime").Each(func(i int, s *goquery.Selection) {
            if i == 0 {
                if t, err := time.Parse(time.RFC3339, strings.TrimSpace(s.Text())); err == nil {
                    lastTime = t
                }
            }
        })
    }
    msgCount := doc.Find(".tgme_widget_message_wrap").Length()
    var texts []string
    doc.Find(".tgme_widget_message_text, pre, code").Each(func(i int, s *goquery.Selection) {
        texts = append(texts, s.Text())
    })
    has := anyConfigInText(strings.Join(texts, "\n"))
    if lastTime.IsZero() {
        return ScanResult{}, fmt.Errorf("no timestamp found")
    }
    daysSince := int(time.Since(lastTime).Hours() / 24)
    status := "inactive"
    if has && daysSince <= *activeDays {
        status = "active"
    }
    safePrintf("[INFO] %s -> last: %s (%d days), config: %v, status: %s\n",
        origURL, lastTime.Format("2006-01-02"), daysSince, has, status)
    return ScanResult{
        URL:          origURL,
        LastPost:     lastTime,
        HasConfig:    has,
        Status:       status,
        MessageCount: msgCount,
        Timestamp:    time.Now(),
    }, nil
}

func anyConfigInText(text string) bool {
    for _, re := range configPatterns {
        if re.MatchString(text) {
            return true
        }
    }
    return false
}

func extractChannelName(rawURL string) string {
    re := regexp.MustCompile(`t\.me/(?:s/)?([^/?]+)`)
    m := re.FindStringSubmatch(rawURL)
    if len(m) > 1 {
        return m[1]
    }
    return ""
}

// ---------- I/O ----------
func readCSV(path string) ([][]string, []string, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, nil, err
    }
    defer f.Close()
    rd := csv.NewReader(f)
    rd.FieldsPerRecord = -1
    all, err := rd.ReadAll()
    if err != nil {
        return nil, nil, err
    }
    if len(all) == 0 {
        return nil, nil, nil
    }
    return all[1:], all[0], nil
}

func writeActiveCSV(path string, headers []string, active []ScanResult) error {
    finalHeaders := make([]string, len(headers))
    copy(finalHeaders, headers)
    statusIdx := -1
    flagIdx := -1
    for i, h := range finalHeaders {
        if strings.EqualFold(h, "Status") {
            statusIdx = i
        }
        if strings.EqualFold(h, "AllMessagesFlag") {
            flagIdx = i
        }
    }
    if statusIdx == -1 {
        finalHeaders = append(finalHeaders, "Status")
        statusIdx = len(finalHeaders) - 1
    }
    if flagIdx == -1 {
        finalHeaders = append(finalHeaders, "AllMessagesFlag")
        flagIdx = len(finalHeaders) - 1
    }
    rows := make([][]string, 0, len(active))
    for _, res := range active {
        row := make([]string, len(finalHeaders))
        row[0] = res.URL
        row[statusIdx] = "active"
        row[flagIdx] = "true"
        rows = append(rows, row)
    }
    outFile, err := os.Create(path)
    if err != nil {
        return err
    }
    defer outFile.Close()
    w := csv.NewWriter(outFile)
    defer w.Flush()
    if err := w.Write(finalHeaders); err != nil {
        return err
    }
    for _, row := range rows {
        if err := w.Write(row); err != nil {
            return err
        }
    }
    return nil
}

func loadDeadArchive(file string) map[string]DeadChannelInfo {
    m := make(map[string]DeadChannelInfo)
    data, err := os.ReadFile(file)
    if err != nil {
        return m
    }
    var list []DeadChannelInfo
    if err := json.Unmarshal(data, &list); err != nil {
        return m
    }
    for _, item := range list {
        m[item.URL] = item
    }
    return m
}

func saveDeadArchive(file string, m map[string]DeadChannelInfo) {
    list := make([]DeadChannelInfo, 0, len(m))
    for _, v := range m {
        list = append(list, v)
    }
    sort.Slice(list, func(i, j int) bool { return list[i].URL < list[j].URL })
    data, _ := json.MarshalIndent(list, "", "  ")
    os.WriteFile(file, data, 0644)
    fmt.Printf("✅ Saved %s with %d entries.\n", file, len(list))
}

func safePrintf(format string, args ...interface{}) {
    printMutex.Lock()
    defer printMutex.Unlock()
    fmt.Printf(format, args...)
}
