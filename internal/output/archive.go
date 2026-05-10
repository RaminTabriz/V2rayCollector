package output

import (
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/RaminTabriz/V2rayCollector/internal/cache"
    "github.com/projectdiscovery/gologger"
)

var (
    archiveMutex    sync.RWMutex
    lastArchiveTime int64
    archiveTimeFile = "last_archive_time.txt"
)

func ArchiveDaily(c *cache.Cache) {
    loadLastArchiveTime()

    today := time.Now().Format("2006-01-02")
    archiveDir := filepath.Join("🗄️ daily_archive", today)
    markerFile := filepath.Join(archiveDir, ".done")

    if _, err := os.Stat(markerFile); err == nil {
        gologger.Debug().Msg("Already archived today, skipping")
        return
    }

    gologger.Info().Msgf("📦 Running daily archive for %s", today)

    if err := os.MkdirAll(archiveDir, 0755); err != nil {
        gologger.Error().Msgf("Failed to create archive dir: %v", err)
        return
    }

    srcDir := "📦 all_configs"
    dstDir := filepath.Join(archiveDir, "📦 all_configs")
    if err := copyDir(srcDir, dstDir); err != nil {
        gologger.Error().Msgf("Failed to copy %s to archive: %v", srcDir, err)
        return
    }

    if err := os.RemoveAll(srcDir); err != nil {
        gologger.Warning().Msgf("Failed to remove %s: %v", srcDir, err)
    }

    if err := os.MkdirAll(srcDir, 0755); err != nil {
        gologger.Warning().Msgf("Failed to recreate %s: %v", srcDir, err)
    }
    os.MkdirAll(filepath.Join(srcDir, "📡 telegram"), 0755)
    os.MkdirAll(filepath.Join(srcDir, "🔗 subscription"), 0755)

    if err := os.WriteFile(markerFile, []byte("archived"), 0644); err != nil {
        gologger.Warning().Msgf("Failed to write marker: %v", err)
    }

    updateLastArchiveTime()
    gologger.Info().Msgf("Archived %s to %s", srcDir, archiveDir)
}

func copyDir(src, dst string) error {
    return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil
        }
        rel, err := filepath.Rel(src, path)
        if err != nil {
            return nil
        }
        destPath := filepath.Join(dst, rel)
        if info.IsDir() {
            return os.MkdirAll(destPath, info.Mode())
        }
        data, err := os.ReadFile(path)
        if err != nil {
            gologger.Warning().Msgf("Failed to read %s: %v", path, err)
            return nil
        }
        if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
            return err
        }
        return os.WriteFile(destPath, data, info.Mode())
    })
}

func loadLastArchiveTime() {
    archiveMutex.RLock()
    defer archiveMutex.RUnlock()
    data, err := os.ReadFile(archiveTimeFile)
    if err != nil {
        lastArchiveTime = 0
        return
    }
    val, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
    if err != nil {
        lastArchiveTime = 0
        return
    }
    lastArchiveTime = val
}

func updateLastArchiveTime() {
    archiveMutex.Lock()
    defer archiveMutex.Unlock()
    lastArchiveTime = time.Now().Unix()
    if err := os.WriteFile(archiveTimeFile, []byte(fmt.Sprintf("%d", lastArchiveTime)), 0644); err != nil {
        gologger.Warning().Msgf("Failed to save archive time: %v", err)
    }
}

func GetLastArchiveTime() int64 {
    archiveMutex.RLock()
    defer archiveMutex.RUnlock()
    return lastArchiveTime
}
