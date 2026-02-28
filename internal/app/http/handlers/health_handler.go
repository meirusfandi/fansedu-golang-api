package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

func Health() http.HandlerFunc {
	type resp struct {
		Status string `json:"status"`
		Time   string `json:"time"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp{
			Status: "ok",
			Time:   time.Now().UTC().Format(time.RFC3339),
		})
	}
}

