package middleware

import (
	"sync"
	"time"
)

// LoginLimiter membatasi percobaan login GAGAL: setelah `max` kegagalan dalam
// `window`, percobaan berikutnya ditolak (429) sampai kegagalan lama kedaluwarsa.
// Implementasi in-memory sederhana, cukup untuk satu instance server.
type LoginLimiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	fails  map[string][]time.Time
}

func NewLoginLimiter(max int, window time.Duration) *LoginLimiter {
	return &LoginLimiter{max: max, window: window, fails: map[string][]time.Time{}}
}

func (l *LoginLimiter) prune(key string, now time.Time) {
	list := l.fails[key]
	keep := list[:0]
	for _, t := range list {
		if now.Sub(t) < l.window {
			keep = append(keep, t)
		}
	}
	if len(keep) == 0 {
		delete(l.fails, key)
		return
	}
	l.fails[key] = keep
}

func (l *LoginLimiter) Blocked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prune(key, time.Now())
	return len(l.fails[key]) >= l.max
}

func (l *LoginLimiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.prune(key, now)
	l.fails[key] = append(l.fails[key], now)

	// Bersihkan key lain yang sudah kedaluwarsa agar map tidak membengkak.
	if len(l.fails) > 10000 {
		for k := range l.fails {
			l.prune(k, now)
		}
	}
}

func (l *LoginLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.fails, key)
}
