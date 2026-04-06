package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// extractPaymentProofUpload membaca multipart form (field proof), menyimpan file, mengembalikan path publik (/uploads/...).
// Jika httpStatus != 0, caller harus memanggil writeError(w, httpStatus, code, msg).
func extractPaymentProofUpload(r *http.Request, orderID string) (proofPath, senderAccountNo, senderName string, httpStatus int, code, msg string) {
	if orderID == "" {
		return "", "", "", http.StatusBadRequest, "bad_request", "orderId required"
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return "", "", "", http.StatusBadRequest, "bad_request", "invalid multipart form"
	}
	file, fh, err := r.FormFile("proof")
	if err != nil {
		return "", "", "", http.StatusBadRequest, "bad_request", "proof file required"
	}
	defer file.Close()
	senderAccountNo = strings.TrimSpace(r.FormValue("senderAccountNo"))
	senderName = strings.TrimSpace(r.FormValue("senderName"))

	filename := "proof.dat"
	if fh != nil && fh.Filename != "" {
		filename = fh.Filename
	}
	safeName := strings.ReplaceAll(filepath.Base(filename), "..", "")
	if safeName == "" {
		safeName = "proof"
	}
	if fh == nil {
		return "", "", "", http.StatusBadRequest, "BAD_REQUEST", "Berkas bukti pembayaran tidak valid."
	}
	if fh.Size > maxPaymentProofBytes {
		return "", "", "", http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "Ukuran file bukti pembayaran melebihi batas."
	}
	ct := strings.ToLower(strings.TrimSpace(fh.Header.Get("Content-Type")))
	if ct != "" {
		if _, ok := allowedPaymentProofContentTypes[ct]; !ok {
			return "", "", "", http.StatusBadRequest, "INVALID_FILE_TYPE", "Tipe file tidak diizinkan. Gunakan JPG, PNG, WebP, atau PDF."
		}
	} else {
		ext := strings.ToLower(filepath.Ext(safeName))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" && ext != ".pdf" {
			return "", "", "", http.StatusBadRequest, "INVALID_FILE_TYPE", "Ekstensi file tidak diizinkan. Gunakan .jpg, .png, .webp, atau .pdf."
		}
	}
	dir := filepath.Join(paymentProofUploadDir, orderID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", "", http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal menyimpan berkas."
	}
	dstPath := filepath.Join(dir, safeName)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", "", "", http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal menyimpan berkas."
	}
	_, _ = io.Copy(dst, file)
	dst.Close()
	webPath := "/" + filepath.ToSlash(dstPath)
	return webPath, senderAccountNo, senderName, 0, "", ""
}
