package main

import (
    "encoding/csv"
    "encoding/json"
    "fmt"
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
    defaultRetryCount   = 3
    defaultBaseDelay    = 1 * time.Second
    defaultJitter       = 500 * time.Millisecond
    activeDays          = 30
    defaultConcurrency  = 5

    dataDir             = "data"
    deadChannelsRecent  = dataDir + "/dead_channels_recent.json"
    deadChannelsOld     = dataDir + "/dead_channels_old.json"
    activeChannelsFile  = "channels.csv"
)

var (
    client = &http.Client{Timeout: 15 * time.Second}

    // الگوهای تشخیص کانفیگ (همانند سایر ابزارها)
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
        regexp.MustCompile(`tg://proxy\?[^\s]+`),
        regexp.MustCompile(`tg://socks\?[^\s]+`),
        regexp.MustCompile(`slipnet://[^\s]+`),
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

type ReviveResult struct {
    URL       string    `json:"url"`
    LastPost  time.Time `json:"last_post"`
    HasConfig bool      `json:"has_config"`
    Status    string    `json:"status"`
    Revived   bool      `json:"revived"`
    Error     string    `json:"error,omitempty"`
}

func main() {
    os.MkdirAll(dataDir, 0755)
    os.MkdirAll("reports", 0755)

    recentDead := loadDeadArchive(deadChannelsRecent)
    oldDead := loadDeadArchive(deadChannelsOld)

    allDead := make(map[string]DeadChannelInfo)
    for k, v := range recentDead {
        allDead[k] = v
    }
    for k, v := range oldDead {
        allDead[k] = v
    }

    if len(allDead) == 0 {
        fmt.Println("Archive is empty. Nothing to revive.")
        return
    }
    fmt.Printf("Loaded %d archived channels (recent: %d, old: %d).\n", len(allDead), len(recentDead), len(oldDead))

    activeMap := loadActiveChannels()
    var revivedList []string
    var results []ReviveResult

    jobs := make(chan string, len(allDead))
    resultsCh := make(chan ReviveResult, len(allDead))
    var wg sync.WaitGroup
    workers := defaultConcurrency
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for url := range jobs {
                res := checkChannelWithRetry(url)
                resultsCh <- res
                // جلوگیری از درخواست بیش از حد
                time.Sleep(time.Duration(rand.Intn(500)+500) * time.Millisecond)
            }
        }()
    }
    for url := range allDead {
        jobs <- url
    }
    close(jobs)
    wg.Wait()
    close(resultsCh)

    updatedRecent := make(map[string]DeadChannelInfo)
    updatedOld := make(map[string]DeadChannelInfo)

    for res := range resultsCh {
        results = append(results, res)
        if res.Revived {
            if !activeMap[res.URL] {
                revivedList = append(revivedList, res.URL)
            }
            continue
        }
        info := DeadChannelInfo{
            URL:       res.URL,
            LastPost:  res.LastPost.Unix(),
            CheckedAt: time.Now().Unix(),
        }
        daysSince := int(time.Since(res.LastPost).Hours() / 24)
        if daysSince > 365 {
            updatedOld[res.URL] = info
        } else {
            updatedRecent[res.URL] = info
        }
    }

    // حفظ کانال‌های اسکن نشده (در صورت وجود)
    for k, v := range recentDead {
        if _, exists := updatedRecent[k]; !exists {
            updatedRecent[k] = v
        }
    }
    for k, v := range oldDead {
        if _, exists := updatedOld[k]; !exists {
            updatedOld[k] = v
        }
    }

    if len(revivedList) > 0 {
        addToActiveChannels(revivedList)
        fmt.Printf("✅ Added %d revived channels to %s\n", len(revivedList), activeChannelsFile)
    } else {
        fmt.Println("No revived channels found.")
    }

    saveDeadArchive(deadChannelsRecent, updatedRecent)
    saveDeadArchive(deadChannelsOld, updatedOld)

    fmt.Printf("✅ Revive scan finished. Revived: %d, Still dead (recent: %d, old: %d)\n",
        len(revivedList), len(updatedRecent), len(updatedOld))
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

func checkChannelWithRetry(url string) ReviveResult {
    var lastErr error
    for attempt := 1; attempt <= defaultRetryCount; attempt++ {
        res, err := analyzeChannelForRevive(url)
        if err == nil {
            return res
        }
        lastErr = err
        delay := defaultBaseDelay * time.Duration(1<<uint(attempt-1))
        jitter := time.Duration(rand.Int63n(int64(defaultJitter)))
        time.Sleep(delay + jitter)
    }
    return ReviveResult{URL: url, Status: "error", Error: lastErr.Error(), Revived: false}
}

func analyzeChannelForRevive(channelURL string) (ReviveResult, error) {
    channelName := extractChannelName(channelURL)
    if channelName == "" {
        return ReviveResult{}, fmt.Errorf("invalid URL")
    }
    // تلاش با RSS
    rssURL := fmt.Sprintf("https://t.me/s/%s.rss", channelName)
    res, err := fetchFromRSSRevive(rssURL, channelURL)
    if err == nil {
        return res, nil
    }
    // در صورت شکست RSS، از HTML استفاده کن
    htmlURL := fmt.Sprintf("https://t.me/s/%s", channelName)
    return fetchFromHTMLRevive(htmlURL, channelURL)
}

func fetchFromRSSRevive(rssURL, origURL string) (ReviveResult, error) {
    resp, err := client.Get(rssURL)
    if err != nil {
        return ReviveResult{}, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return ReviveResult{}, fmt.Errorf("HTTP %d", resp.StatusCode)
    }
    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return ReviveResult{}, err
    }
    var latestTime time.Time
    var anyConfig bool
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
        return ReviveResult{}, fmt.Errorf("no pubDate")
    }
    daysSince := int(time.Since(latestTime).Hours() / 24)
    revived := anyConfig && daysSince <= activeDays
    status := "inactive"
    if revived {
        status = "active"
    }
    safePrintf("[INFO] %s -> last: %s (%d days), config: %v, revived: %v\n",
        origURL, latestTime.Format("2006-01-02"), daysSince, anyConfig, revived)
    return ReviveResult{
        URL:       origURL,
        LastPost:  latestTime,
        HasConfig: anyConfig,
        Status:    status,
        Revived:   revived,
    }, nil
}

func fetchFromHTMLRevive(htmlURL, origURL string) (ReviveResult, error) {
    resp, err := client.Get(htmlURL)
    if err != nil {
        return ReviveResult{}, err
    }
    defer resp.Body.Close()
    if resp.StatusCode == 404 {
        return ReviveResult{URL: origURL, Status: "banned", Revived: false, Error: "not found"}, nil
    }
    if resp.StatusCode != 200 {
        return ReviveResult{}, fmt.Errorf("HTTP %d", resp.StatusCode)
    }
    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return ReviveResult{}, err
    }
    var lastTime time.Time
    doc.Find("time").Each(func(i int, s *goquery.Selection) {
        if i == 0 && lastTime.IsZero() {
            if dt, ok := s.Attr("datetime"); ok {
                t, _ := time.Parse(time.RFC3339, dt)
                lastTime = t
            }
        }
    })
    if lastTime.IsZero() {
        doc.Find(".datetime").Each(func(i int, s *goquery.Selection) {
            if i == 0 && lastTime.IsZero() {
                t, _ := time.Parse(time.RFC3339, strings.TrimSpace(s.Text()))
                lastTime = t
            }
        })
    }
    var texts []string
    doc.Find(".tgme_widget_message_text, pre, code").Each(func(i int, s *goquery.Selection) {
        texts = append(texts, s.Text())
    })
    has := anyConfigInText(strings.Join(texts, "\n"))
    if lastTime.IsZero() {
        return ReviveResult{}, fmt.Errorf("no timestamp found")
    }
    daysSince := int(time.Since(lastTime).Hours() / 24)
    revived := has && daysSince <= activeDays
    status := "inactive"
    if revived {
        status = "active"
    }
    safePrintf("[INFO] %s -> last: %s (%d days), config: %v, revived: %v\n",
        origURL, lastTime.Format("2006-01-02"), daysSince, has, revived)
    return ReviveResult{
        URL:       origURL,
        LastPost:  lastTime,
        HasConfig: has,
        Status:    status,
        Revived:   revived,
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

// ---------- توابع کمکی I/O ----------
func loadActiveChannels() map[string]bool {
    m := make(map[string]bool)
    f, err := os.Open(activeChannelsFile)
    if err != nil {
        return m
    }
    defer f.Close()
    r := csv.NewReader(f)
    records, err := r.ReadAll()
    if err != nil {
        return m
    }
    if len(records) < 2 {
        return m
    }
    for _, row := range records[1:] {
        if len(row) > 0 {
            m[row[0]] = true
        }
    }
    return m
}

func addToActiveChannels(urls []string) {
    records, headers, err := readCSV(activeChannelsFile)
    if err != nil {
        fmt.Printf("Error reading CSV: %v\n", err)
        return
    }
    activeSet := make(map[string]bool)
    for _, row := range records {
        if len(row) > 0 {
            activeSet[row[0]] = true
        }
    }
    for _, url := range urls {
        if !activeSet[url] {
            records = append(records, []string{url, "false"})
            fmt.Printf("Adding revived channel: %s\n", url)
        }
    }
    if err := writeCSV(activeChannelsFile, headers, records); err != nil {
        fmt.Printf("Error writing CSV: %v\n", err)
    }
}

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

func writeCSV(path string, headers []string, records [][]string) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()
    w := csv.NewWriter(f)
    defer w.Flush()
    if err := w.Write(headers); err != nil {
        return err
    }
    for _, row := range records {
        if err := w.Write(row); err != nil {
            return err
        }
    }
    return nil
}

func safePrintf(format string, args ...interface{}) {
    printMutex.Lock()
    defer printMutex.Unlock()
    fmt.Printf(format, args...)
}
