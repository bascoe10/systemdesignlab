// LEVEL 5 SKELETON — api-gateway
//
// Contract: system/contracts/api-gateway.yaml. Rate-limit per client IP,
// proxy POST /api/shorten to the shortener and GET /r/{code} to the
// redirector. Upstream URLs and rate limits come from config.yaml
// (internal/config is available).
//
// Everything below boots and exposes metrics so the stack starts before
// you've written a line. Replace the 501s with the real thing.
package main

import (
	"log"
	"net/http"

	"github.com/systemdesignlab/url-shortener/internal/obs"
)

func notImplemented(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented — see BRIEFING.md", http.StatusNotImplemented)
}

func main() {
	// TODO: cfg, err := config.Load()
	// TODO: build reverse proxies to cfg.Services.ShortenerURL / RedirectorURL
	//       (net/http/httputil has what you need)
	// TODO: per-IP rate limiting (golang.org/x/time/rate is in go.mod);
	//       count rejections with obs.RateLimitRejected

	mux := http.NewServeMux()
	mux.Handle("/api/shorten", obs.Instrument("shorten", http.HandlerFunc(notImplemented)))
	mux.Handle("/r/", obs.Instrument("redirect", http.HandlerFunc(notImplemented)))
	mux.Handle("/metrics", obs.MetricsHandler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })

	log.Println("api-gateway (skeleton) listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
