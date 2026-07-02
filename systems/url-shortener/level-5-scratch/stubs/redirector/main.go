// LEVEL 5 SKELETON — redirector (read-heavy path)
//
// Contract: system/contracts/redirector.yaml. Cache-aside: try the cache
// node your ring picks; on miss, read Postgres, backfill the cache
// asynchronously, and 302 either way. Unknown code → 404.
//
// The resilience requirement is the interesting part: when cache nodes are
// down (make chaos-kill-cache), this service must keep serving from the
// database — degraded latency, zero errors.
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
	// TODO: load config, open the store, build the cache cluster.
	// TODO: GET /r/{code} —
	//   1. cache lookup via the ring (a ring error means: bypass, not fail)
	//   2. hit → 302 immediately
	//   3. miss → Postgres; unknown → 404; found → 302 + async backfill
	//      (why async? what does a synchronous backfill cost you at p99?)
	// TODO: make /healthz return 503 when the database is unreachable.

	mux := http.NewServeMux()
	mux.Handle("/r/", obs.Instrument("redirect", http.HandlerFunc(notImplemented)))
	mux.Handle("/metrics", obs.MetricsHandler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })

	log.Println("redirector (skeleton) listening on :8082")
	log.Fatal(http.ListenAndServe(":8082", mux))
}
