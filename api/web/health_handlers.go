package web

import (
	"encoding/json"
	"net/http"

	"github.com/masudur-rahman/expense-tracker-bot/configs"
)

func HandleHealthz(w http.ResponseWriter, _ *http.Request) {
	resp := map[string]string{"status": "ok", "db": "ok"}

	if err := configs.PingDatabase(); err != nil {
		resp["status"] = "degraded"
		resp["db"] = err.Error()
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
