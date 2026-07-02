// The shortener owns the write path: generate a short code, persist it,
// then write through to the cache node chosen by the consistent hash ring.
package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/systemdesignlab/url-shortener/internal/cache"
	"github.com/systemdesignlab/url-shortener/internal/config"
	"github.com/systemdesignlab/url-shortener/internal/obs"
	"github.com/systemdesignlab/url-shortener/internal/store"
)

const codeAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const codeLength = 7

func newCode() (string, error) {
	buf := make([]byte, codeLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i, b := range buf {
		buf[i] = codeAlphabet[int(b)%len(codeAlphabet)]
	}
	return string(buf), nil
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

type server struct {
	store   *store.Store
	cluster *cache.Cluster
	baseURL string
}

func (s *server) handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	parsed, err := url.ParseRequestURI(req.URL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		http.Error(w, "url must be absolute http(s)", http.StatusBadRequest)
		return
	}

	var code string
	for attempt := 0; attempt < 3; attempt++ {
		code, err = newCode()
		if err != nil {
			break
		}
		err = s.store.Insert(r.Context(), code, req.URL)
		if !errors.Is(err, store.ErrConflict) {
			break
		}
	}
	if err != nil {
		log.Printf("shorten: %v", err)
		http.Error(w, "could not create short URL", http.StatusInternalServerError)
		return
	}

	// Write-through: new URLs are read soon after creation, so cache now.
	s.cluster.Set(r.Context(), code, req.URL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(shortenResponse{Code: code, ShortURL: s.baseURL + "/r/" + code})
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
	cache.StartMaintenance(cfg, cluster.Provider)

	baseURL := os.Getenv("PUBLIC_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	srv := &server{store: st, cluster: cluster, baseURL: baseURL}

	mux := http.NewServeMux()
	mux.Handle("/api/shorten", obs.Instrument("shorten", http.HandlerFunc(srv.handleShorten)))
	mux.Handle("/metrics", obs.MetricsHandler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if !st.Healthy(context.Background()) {
			http.Error(w, "db unreachable", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})

	log.Println("shortener listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
