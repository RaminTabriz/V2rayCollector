// در داخل struct Subscription اضافه کنید:
type Subscription struct {
    urls        []string
    concurrency int
    successCount int
    failedCount  int
    mu           sync.Mutex
}

// سپس در متد worker، هنگام موفقیت یا شکست:
func (s *Subscription) worker(...) {
    // ...
    if err == nil && resp.StatusCode == 200 {
        s.mu.Lock()
        s.successCount++
        s.mu.Unlock()
    } else {
        s.mu.Lock()
        s.failedCount++
        s.mu.Unlock()
    }
}

// و متدهای دسترسی:
func (s *Subscription) GetSuccessCount() int {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.successCount
}
func (s *Subscription) GetFailedCount() int {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.failedCount
}
