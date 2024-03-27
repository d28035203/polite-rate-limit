// polite-rate-limit — token-bucket rate limiter in front of a tiny HTTP API
package main

import (
	"encoding/json"
	"log"
	"net"
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

func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func withLimit(b *Bucket, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !b.Allow(clientKey(r)) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded"})
			return
		}
		next(w, r)
	}
}

func main() {
	// 5 tokens/sec refill, burst 10
	lim := NewBucket(5, 10)
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", withLimit(lim, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"msg": "hello"})
	}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
