// The api-gateway is the single public entrypoint: it rate-limits per
// client IP and proxies to the shortener (writes) and redirector (reads).
package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/systemdesignlab/url-shortener/internal/config"
	"github.com/systemdesignlab/url-shortener/internal/obs"
)

// ipLimiter hands out one token bucket per client IP and evicts idle ones.
type ipLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucketEntry
	rps     rate.Limit
	burst   int
}

type bucketEntry struct {
	lim  *rate.Limiter
	seen time.Time
}

func newIPLimiter(rps, burst int) *ipLimiter {
	l := &ipLimiter{buckets: map[string]*bucketEntry{}, rps: rate.Limit(rps), burst: burst}
	go func() {
		for range time.Tick(time.Minute) {
			l.mu.Lock()
			for ip, e := range l.buckets {
				if time.Since(e.seen) > 3*time.Minute {
					delete(l.buckets, ip)
				}
			}
			l.mu.Unlock()
		}
	}()
	return l
}

func (l *ipLimiter) allow(remoteAddr string) bool {
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		ip = remoteAddr
	}
	l.mu.Lock()
	e, ok := l.buckets[ip]
	if !ok {
		e = &bucketEntry{lim: rate.NewLimiter(l.rps, l.burst)}
		l.buckets[ip] = e
	}
	e.seen = time.Now()
	l.mu.Unlock()
	return e.lim.Allow()
}

func mustProxy(raw string) *httputil.ReverseProxy {
	u, err := url.Parse(raw)
	if err != nil {
		log.Fatalf("bad upstream URL %q: %v", raw, err)
	}
	p := httputil.NewSingleHostReverseProxy(u)
	p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy %s: %v", r.URL.Path, err)
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}
	return p
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	limiter := newIPLimiter(cfg.Gateway.RateLimitRPS, cfg.Gateway.Burst)
	shortener := mustProxy(cfg.Services.ShortenerURL)
	redirector := mustProxy(cfg.Services.RedirectorURL)

	limited := func(route string, next http.Handler) http.Handler {
		return obs.Instrument(route, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.allow(r.RemoteAddr) {
				obs.RateLimitRejected.Inc()
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		}))
	}

	mux := http.NewServeMux()
	mux.Handle("/api/shorten", limited("shorten", shortener))
	mux.Handle("/r/", limited("redirect", redirector))
	mux.Handle("/metrics", obs.MetricsHandler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })

	log.Println("api-gateway listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
