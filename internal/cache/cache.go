package cache

import (
    "encoding/json"
    "os"
    "sync"
    "time"
)

type Entry struct {
    Timestamp   int64  `json:"timestamp"`
    Source      string `json:"source"`
    Fingerprint string `json:"fingerprint"`
    Channel     string `json:"channel,omitempty"`
    Protocol    string `json:"protocol"`
}

type Cache struct {
    mu    sync.RWMutex
    path  string
    items map[string]Entry
    fpMap map[string]string // fingerprint -> config key
}

func New(path string) *Cache {
    return &Cache{
        path:  path,
        items: make(map[string]Entry),
        fpMap: make(map[string]string),
    }
}

func (c *Cache) Load() {
    data, err := os.ReadFile(c.path)
    if err != nil {
        return
    }
    var cd struct {
        Configs map[string]Entry `json:"configs"`
    }
    if err := json.Unmarshal(data, &cd); err != nil {
        return
    }
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items = cd.Configs
    for key, ent := range c.items {
        if ent.Fingerprint != "" {
            c.fpMap[ent.Fingerprint] = key
        }
    }
}

func (c *Cache) Add(config, source, channel, protocol string) bool {
    fp := ComputeFingerprint(config, protocol)
    c.mu.Lock()
    defer c.mu.Unlock()
    if existingKey, ok := c.fpMap[fp]; ok {
        _ = existingKey
        return false // duplicate
    }
    c.items[config] = Entry{
        Timestamp:   time.Now().Unix(),
        Source:      source,
        Fingerprint: fp,
        Channel:     channel,
        Protocol:    protocol,
    }
    c.fpMap[fp] = config
    return true
}

func (c *Cache) GetAll() map[string]Entry {
    c.mu.RLock()
    defer c.mu.RUnlock()
    out := make(map[string]Entry)
    for k, v := range c.items {
        out[k] = v
    }
    return out
}

func (c *Cache) GetBySource(source string) map[string]Entry {
    c.mu.RLock()
    defer c.mu.RUnlock()
    out := make(map[string]Entry)
    for k, v := range c.items {
        if v.Source == source {
            out[k] = v
        }
    }
    return out
}

func (c *Cache) Save() {
    c.mu.RLock()
    defer c.mu.RUnlock()
    data, err := json.MarshalIndent(struct {
        Configs map[string]Entry `json:"configs"`
    }{c.items}, "", "  ")
    if err != nil {
        return
    }
    os.WriteFile(c.path, data, 0644)
}
