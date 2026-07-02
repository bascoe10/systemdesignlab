// The redirector owns the read-heavy path: cache-aside lookup with database
// fallback and asynchronous cache backfill on miss.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/systemdesignlab/url-shortener/internal/cache"
	"github.com/systemdesignlab/url-shortener/internal/config"
	"github.com/systemdesignlab/url-shortener/internal/obs"
	"github.com/systemdesignlab/url-shortener/internal/store"
)

type server struct {
	store   *store.Store
	cluster *cache.Cluster
}

func (s *server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/r/")
	if code == "" || strings.Contains(code, "/") {
		http.Error(w, "bad short code", http.StatusBadRequest)
		return
	}

	if target, ok := s.cluster.Get(r.Context(), code); ok {
		http.Redirect(w, r, target, http.StatusFound)
		return
	}

	target, err := s.store.Get(r.Context(), code)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("redirect %s: %v", code, err)
		http.Error(w, "lookup failed", http.StatusInternalServerError)
		return
	}

	// Backfill asynchronously so the user never waits on the cache write.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		s.cluster.Set(ctx, code, target)
	}()

	http.Redirect(w, r, target, http.StatusFound)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	st, err := store.Open(cfg)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	cluster, err := cache.NewClusterFromConfig(cfg)
	if err != nil {
		log.Fatalf("cache: %v", err)
	}
	srv := &server{store: st, cluster: cluster}

	mux := http.NewServeMux()
	mux.Handle("/r/", obs.Instrument("redirect", http.HandlerFunc(srv.handleRedirect)))
	mux.Handle("/metrics", obs.MetricsHandler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if !st.Healthy(context.Background()) {
			http.Error(w, "db unreachable", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})

	log.Println("redirector listening on :8082")
	log.Fatal(http.ListenAndServe(":8082", mux))
}
