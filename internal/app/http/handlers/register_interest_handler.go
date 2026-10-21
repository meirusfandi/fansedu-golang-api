package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
)

// Request struct for register interest
// Sesuai payload JSON yang diminta
// untuk validasi dan decoding body
 type RegisterInterestRequest struct {
	NamaOrangTua string `json:"namaOrangTua"`
	Whatsapp     string `json:"whatsapp"`
	NamaAnak     string `json:"namaAnak"`
	UsiaAnak     string `json:"usiaAnak"`
	ProgramMinat string `json:"programMinat"`
	Catatan      string `json:"catatan"`
}

// Response struct
 type RegisterInterestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    *struct {
		NamaAnak     string `json:"namaAnak"`
		ProgramMinat string `json:"programMinat"`
	} `json:"data,omitempty"`
}

func validWhatsappNumber(number string) bool {
	re := regexp.MustCompile(`^\d{9,15}$`)
	return re.MatchString(number)
}

func (h *Deps) RegisterInterestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req RegisterInterestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleErrorResponse(w, http.StatusBadRequest, "Request body tidak valid")
		return
	}

	// Validasi input wajib
	if req.NamaOrangTua == "" || req.Whatsapp == "" || req.NamaAnak == "" || req.UsiaAnak == "" {
		handleErrorResponse(w, http.StatusBadRequest, "Mohon isi semua field wajib bertanda bintang.")
		return
	}

	// Validasi nomor whatsapp
	if !validWhatsappNumber(req.Whatsapp) {
		handleErrorResponse(w, http.StatusBadRequest, "Nomor WhatsApp tidak valid. Harus berupa angka dengan 9-15 digit.")
		return
	}

	// Validasi usia anak positif
	usia, err := strconv.Atoi(req.UsiaAnak)
	if err != nil || usia <= 0 {
		handleErrorResponse(w, http.StatusBadRequest, "Usia anak harus angka positif.")
		return
	}

	// TODO: Simulasi penyimpanan data (bisa dibuat repo baru dan panggil disini)

	resp := RegisterInterestResponse{
		Success: true,
		Message: "Status pendaftaran minat berhasil disimpan terenkripsi.",
		Data: &struct {
			NamaAnak     string `json:"namaAnak"`
			ProgramMinat string `json:"programMinat"`
		}{req.NamaAnak, req.ProgramMinat},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return
	}
}

func handleErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(RegisterInterestResponse{
		Success: false,
		Message: message,
	})
}
