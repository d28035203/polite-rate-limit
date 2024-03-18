// polite-rate-limit — token bucket middleware demo
package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Bucket struct {
	mu     sync.Mutex
	tokens map[string]float64
	last   map[string]time.Time
	rate   float64
	burst  float64
}

func NewBucket(rate, burst float64) *Bucket {
	return &Bucket{
		tokens: make(map[string]float64),
		last:   make(map[string]time.Time),
		rate:   rate,
		burst:  burst,
	}
}

func (b *Bucket) Allow(key string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	t, ok := b.tokens[key]
	if !ok {
		b.tokens[key] = b.burst - 1
		b.last[key] = now
		return true
	}
	elapsed := now.Sub(b.last[key]).Seconds()
	t += elapsed * b.rate
	if t > b.burst {
		t = b.burst
	}
	b.last[key] = now
	if t < 1 {
		b.tokens[key] = t
		return false
	}
	b.tokens[key] = t - 1
	return true
}

func main() {
	lim := NewBucket(5, 10)
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		if !lim.Allow(r.RemoteAddr) {
			w.Header().Set("Retry-After", "1")
			http.Error(w, `{"error":"slow down, friend"}`, http.StatusTooManyRequests)
			return
		}
		fmt.Fprintln(w, `{"msg":"hello, politely"}`)
	})
	fmt.Println("listening on :8080")
	http.ListenAndServe(":8080", mux)
}
