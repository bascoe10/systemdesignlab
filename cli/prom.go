package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// promQuery runs an instant PromQL query and returns the first sample value.
func promQuery(query string) (float64, error) {
	base := os.Getenv("PROMETHEUS_URL")
	if base == "" {
		base = "http://localhost:9090"
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(base + "/api/v1/query?query=" + url.QueryEscape(query))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var body struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Value [2]any `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, err
	}
	if body.Status != "success" {
		return 0, fmt.Errorf("prometheus returned status %q", body.Status)
	}
	if len(body.Data.Result) == 0 {
		return 0, fmt.Errorf("no data (has a load test run recently?)")
	}
	s, ok := body.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("unexpected value type")
	}
	return strconv.ParseFloat(s, 64)
}

func promReachable() bool {
	_, err := promQuery("vector(1)")
	return err == nil
}
