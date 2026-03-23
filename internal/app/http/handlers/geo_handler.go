package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

// GeoProvinces GET /api/v1/geo/provinces — JSON array [{id,name},...] (emsifa-compatible).
func GeoProvinces(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps == nil || deps.GeoService == nil {
			http.Error(w, `{"error":"geo_unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		body, err := deps.GeoService.ProvincesJSON(r.Context())
		if err != nil {
			http.Error(w, `{"error":"upstream_failed"}`, http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}
}

// GeoRegencies GET /api/v1/geo/regencies/{provinceId}
func GeoRegencies(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps == nil || deps.GeoService == nil {
			http.Error(w, `{"error":"geo_unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		provinceID := chi.URLParam(r, "provinceId")
		body, err := deps.GeoService.RegenciesJSON(r.Context(), provinceID)
		if err != nil {
			if errors.Is(err, service.ErrGeoNotFound) {
				http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error":"upstream_failed"}`, http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}
}
