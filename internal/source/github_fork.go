package source

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/ramin00542/GO_V2rayCollector/internal/fetcher"
    "github.com/projectdiscovery/gologger"
)

type GitHubFork struct {
    targetRepo string
}

func NewGitHubFork(repo string) *GitHubFork {
    return &GitHubFork{targetRepo: repo}
}

func (g *GitHubFork) Scan(ctx context.Context, onFound func(rawURL string)) {
    gologger.Info().Msg("Scanning GitHub forks for subscription links...")
    page := 1
    for {
        select {
        case <-ctx.Done():
            return
        default:
        }
        apiURL := fmt.Sprintf("https://api.github.com/repos/%s/forks?per_page=100&page=%d", g.targetRepo, page)
        req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
        if err != nil {
            break
        }
        req.Header.Set("Accept", "application/vnd.github.v3+json")
        resp, err := fetcher.Client.Do(req)
        if err != nil || resp.StatusCode != 200 {
            if resp != nil {
                if resp.StatusCode == 403 {
                    gologger.Warning().Msg("GitHub API rate limit reached, stopping fork scan")
                }
                resp.Body.Close()
            }
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
            fullName, _ := fork["full_name"].(string)
            if fullName == "" {
                continue
            }
            // Common subscription file paths
            paths := []string{
                "sub.txt", "sub_1.txt", "sub_2.txt", "mixed.txt", "config.txt",
                "v2ray.txt", "vmess.txt", "subscription.txt", "iran.txt",
            }
            for _, p := range paths {
                raw := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s", fullName, p)
                headReq, err := http.NewRequestWithContext(ctx, "HEAD", raw, nil)
                if err != nil {
                    continue
                }
                headResp, err := fetcher.Client.Do(headReq)
                if err == nil && headResp.StatusCode == 200 {
                    onFound(raw)
                    gologger.Info().Msgf("Found subscription in fork %s: %s", fullName, raw)
                }
                if headResp != nil {
                    headResp.Body.Close()
                }
                time.Sleep(200 * time.Millisecond)
            }
        }
        page++
        time.Sleep(1 * time.Second)
    }
}
