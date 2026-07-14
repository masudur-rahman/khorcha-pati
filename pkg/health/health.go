// Package health exposes a minimal HTTP health-check endpoint.
// Wire it into a lightweight http.ServeMux running on port 6336
// so Railway, Docker HEALTHCHECK, and any monitoring system can
// verify the process is alive and initialized.
//
// In main.go:
//
//	go func() {
//	    mux := http.NewServeMux()
//	    mux.Handle("/health", health.Handler(cmd.Version))
//	    log.Fatal(http.ListenAndServe(":6336", mux))
//	}()
package health

import (
	"encoding/json"
	"net/http"
	"time"
)

type response struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

// Handler returns an http.HandlerFunc that responds with JSON {"status":"ok"}.
func Handler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response{
			Status:    "ok",
			Version:   version,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}
}
