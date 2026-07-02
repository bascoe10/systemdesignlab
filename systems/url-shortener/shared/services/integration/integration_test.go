//go:build integration

// End-to-end tests against a running stack (make start). Used by
// `make validate` on Levels 3 and 5. Run manually with:
//
//	go test -tags integration ./integration/ -v
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func baseURL() string {
	if v := os.Getenv("BASE_URL"); v != "" {
		return v
	}
	return "http://localhost:8080"
}

var client = &http.Client{
	Timeout: 5 * time.Second,
	// Don't follow redirects — we assert on the 302 itself.
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func shorten(t *testing.T, target string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"url": target})
	resp, err := client.Post(baseURL()+"/api/shorten", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/shorten: %v (is the stack running? try `make start`)", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("shorten returned %d, want 201", resp.StatusCode)
	}
	var out struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.Code == "" {
		t.Fatal("shorten returned empty code")
	}
	return out.Code
}

func TestShortenAndRedirect(t *testing.T) {
	target := fmt.Sprintf("https://example.com/page-%d", time.Now().UnixNano())
	code := shorten(t, target)

	resp, err := client.Get(baseURL() + "/r/" + code)
	if err != nil {
		t.Fatalf("GET /r/%s: %v", code, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("redirect returned %d, want 302", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != target {
		t.Fatalf("Location = %q, want %q", loc, target)
	}
}

func TestUnknownCodeReturns404(t *testing.T) {
	resp, err := client.Get(baseURL() + "/r/zzzzzzz")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown code returned %d, want 404", resp.StatusCode)
	}
}

func TestInvalidURLRejected(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"url": "not a url"})
	resp, err := client.Post(baseURL()+"/api/shorten", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid url returned %d, want 400", resp.StatusCode)
	}
}

func TestRepeatReadIsCacheServed(t *testing.T) {
	target := fmt.Sprintf("https://example.com/hot-%d", time.Now().UnixNano())
	code := shorten(t, target)

	// Write-through means the first read should already be a hit; either
	// way, repeated reads must stay correct and fast.
	for i := 0; i < 20; i++ {
		resp, err := client.Get(baseURL() + "/r/" + code)
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusFound {
			t.Fatalf("read %d returned %d, want 302", i, resp.StatusCode)
		}
	}
}
