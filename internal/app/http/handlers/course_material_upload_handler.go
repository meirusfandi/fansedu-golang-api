package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const courseMaterialUploadDir = "uploads/course-materials"

const maxCourseMaterialBytes = 40 << 20 // 40 MiB

var allowedCourseMaterialContentTypes = map[string]struct{}{
	"application/pdf":    {},
	"application/msword": {},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   {},
	"application/vnd.ms-powerpoint":                                             {},
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": {},
	"application/octet-stream":                                                  {},
	"application/zip":                                                           {}, // beberapa klien mengirim .pptx sebagai zip
}

func isAllowedCourseMaterialFilename(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".pdf", ".doc", ".docx", ".ppt", ".pptx":
		return true
	default:
		return false
	}
}

// saveCourseMaterialFile menyimpan multipart field "file", mengembalikan path publik "/uploads/course-materials/..."
func saveCourseMaterialFile(w http.ResponseWriter, r *http.Request) (publicPath string, ok bool) {
	if err := r.ParseMultipartForm(maxCourseMaterialBytes + (1 << 20)); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid multipart form")
		return "", false
	}
	file, fh, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "file required (form field: file)")
		return "", false
	}
	defer file.Close()
	if fh == nil || fh.Size == 0 {
		writeError(w, http.StatusBadRequest, "bad_request", "empty file")
		return "", false
	}
	if fh.Size > maxCourseMaterialBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "file_too_large", "file exceeds size limit (40MB)")
		return "", false
	}
	safeName := strings.ReplaceAll(filepath.Base(fh.Filename), "..", "")
	if safeName == "" || !isAllowedCourseMaterialFilename(safeName) {
		writeError(w, http.StatusBadRequest, "invalid_file_type", "only .pdf, .doc, .docx, .ppt, and .pptx allowed")
		return "", false
	}
	ct := strings.ToLower(strings.TrimSpace(fh.Header.Get("Content-Type")))
	if ct != "" {
		if _, ok := allowedCourseMaterialContentTypes[ct]; !ok {
			writeError(w, http.StatusBadRequest, "invalid_file_type", "content-type not allowed for course material upload")
			return "", false
		}
	}
	if err := os.MkdirAll(courseMaterialUploadDir, 0755); err != nil {
		writeInternalError(w, r, err)
		return "", false
	}
	stored := uuid.New().String() + "_" + safeName
	dstPath := filepath.Join(courseMaterialUploadDir, stored)
	dst, err := os.Create(dstPath)
	if err != nil {
		writeInternalError(w, r, err)
		return "", false
	}
	if _, err := io.Copy(dst, file); err != nil {
		_ = dst.Close()
		writeInternalError(w, r, err)
		return "", false
	}
	_ = dst.Close()
	return "/" + filepath.ToSlash(filepath.Join(courseMaterialUploadDir, stored)), true
}

// AdminCourseMaterialUpload POST /api/v1/admin/upload/course-material — multipart file (PDF/DOC/DOCX/PPT/PPTX).
func AdminCourseMaterialUpload(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = deps
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "POST only")
			return
		}
		publicPath, ok := saveCourseMaterialFile(w, r)
		if !ok {
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"url": publicPath})
	}
}

// TrainerCourseMaterialUpload POST /api/v1/trainer/upload/course-material — sama seperti admin upload (PDF/DOC/DOCX/PPT/PPTX).
func TrainerCourseMaterialUpload(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = deps
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "POST only")
			return
		}
		publicPath, ok := saveCourseMaterialFile(w, r)
		if !ok {
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"url": publicPath})
	}
}
