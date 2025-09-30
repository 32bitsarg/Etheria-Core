package handlers

import (
	"encoding/json"
	"net/http"
	"server-backend/config"
	"time"
)

func ServerTimeHandler(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig()
	if err != nil {
		http.Error(w, "Error cargando configuración", http.StatusInternalServerError)
		return
	}
	loc, err := time.LoadLocation(cfg.TimeZone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"server_time": now.Format(time.RFC3339),
	})
}
