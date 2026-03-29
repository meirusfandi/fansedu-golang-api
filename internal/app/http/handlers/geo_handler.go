package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/jsonerror"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

// GeoProvinces GET /api/v1/geo/provinces — JSON array [{id,name},...] (emsifa-compatible).
func GeoProvinces(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps == nil || deps.GeoService == nil {
			jsonerror.Write(w, http.StatusServiceUnavailable, "GEO_UNAVAILABLE", "Data wilayah tidak tersedia sementara.")
			return
		}
		body, err := deps.GeoService.ProvincesJSON(r.Context())
		if err != nil {
			jsonerror.Write(w, http.StatusBadGateway, "GEO_UPSTREAM_ERROR", "Gagal memuat data wilayah.")
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
			jsonerror.Write(w, http.StatusServiceUnavailable, "GEO_UNAVAILABLE", "Data wilayah tidak tersedia sementara.")
			return
		}
		provinceID := chi.URLParam(r, "provinceId")
		body, err := deps.GeoService.RegenciesJSON(r.Context(), provinceID)
		if err != nil {
			if errors.Is(err, service.ErrGeoNotFound) {
				jsonerror.Write(w, http.StatusNotFound, "NOT_FOUND", "Provinsi tidak ditemukan.")
				return
			}
			jsonerror.Write(w, http.StatusBadGateway, "GEO_UPSTREAM_ERROR", "Gagal memuat data wilayah.")
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}
}
