package source

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/RaminTabriz/V2rayCollector/internal/fetcher"
    "github.com/projectdiscovery/gologger"
)

// GitHubFork اسکنر فورک‌های گیت‌هاب برای یافتن فایل‌های ساب‌اسکریپشن
type GitHubFork struct {
    targetRepo string
    perPage    int
}

// NewGitHubFork سازنده جدید برای اسکنر فورک
func NewGitHubFork(repo string) *GitHubFork {
    return &GitHubFork{
        targetRepo: repo,
        perPage:    100,
    }
}

// Scan اسکن فورک‌ها و فراخوانی onFound برای هر لینک ساب‌اسکریپشن پیدا شده
func (g *GitHubFork) Scan(ctx context.Context, onFound func(rawURL string)) {
    gologger.Info().Msg("Scanning GitHub forks for subscription links...")
    page := 1
    found := 0

    for {
        select {
        case <-ctx.Done():
            return
        default:
        }

        apiURL := fmt.Sprintf("https://api.github.com/repos/%s/forks?per_page=%d&page=%d",
            g.targetRepo, g.perPage, page)

        req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
        if err != nil {
            gologger.Warning().Msgf("Failed to create request: %v", err)
            break
        }
        req.Header.Set("Accept", "application/vnd.github.v3+json")

        resp, err := fetcher.Client.Do(req)
        if err != nil {
            gologger.Warning().Msgf("GitHub API request failed: %v", err)
            break
        }

        if resp.StatusCode != 200 {
            if resp.StatusCode == 403 {
                gologger.Warning().Msg("GitHub API rate limit reached, stopping fork scan")
            }
            resp.Body.Close()
            break
        }

        var forks []map[string]interface{}
        if err := json.NewDecoder(resp.Body).Decode(&forks); err != nil {
            resp.Body.Close()
            break
        }
        resp.Body.Close()

        if len(forks) == 0 {
            break
        }

        for _, fork := range forks {
            fullName, ok := fork["full_name"].(string)
            if !ok || fullName == "" {
                continue
            }

            // مسیرهای رایج برای فایل‌های ساب‌اسکریپشن
            paths := []string{
                "sub.txt", "sub_1.txt", "sub_2.txt", "sub_3.txt", "sub_4.txt",
                "mixed.txt", "config.txt", "v2ray.txt", "vmess.txt",
                "subscription.txt", "iran.txt", "all_configs.txt",
                "mci/sub_1.txt", "mci/sub_2.txt", "mtn/sub_1.txt", "mtn/sub_2.txt",
                "xray_final.txt", "sub-link.txt", "clash.yaml",
            }

            for _, p := range paths {
                rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s", fullName, p)

                // HEAD request برای بررسی وجود فایل
                headReq, err := http.NewRequestWithContext(ctx, "HEAD", rawURL, nil)
                if err != nil {
                    continue
                }
                headResp, err := fetcher.Client.Do(headReq)
                if err == nil && headResp.StatusCode == 200 {
                    onFound(rawURL)
                    gologger.Info().Msgf("Found subscription in fork %s: %s", fullName, rawURL)
                    found++
                }
                if headResp != nil {
                    headResp.Body.Close()
                }
                // جلوگیری از درخواست بیش از حد
                time.Sleep(200 * time.Millisecond)
            }
        }

        page++
        time.Sleep(1 * time.Second) // احترام به محدودیت نرخ گیت‌هاب
    }

    gologger.Info().Msgf("Fork scan completed, found %d potential subscription URLs", found)
}
