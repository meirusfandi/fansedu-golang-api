package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/cache"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

func adminCourseToDTO(c domain.Course) dto.CourseResponse {
	tt := c.TrackType
	if tt == "" {
		tt = domain.CourseTrackMeetings
	}
	return dto.CourseResponse{
		ID:          c.ID,
		Title:       c.Title,
		Slug:        c.Slug,
		Description: c.Description,
		Status:      normalizeCourseStatus(c.Status),
		Price:       c.Price,
		Thumbnail:   c.Thumbnail,
		SubjectID:   c.SubjectID,
		CreatedBy:   c.CreatedBy,
		TrackType:   tt,
	}
}

// AdminCourseManageGet GET /api/v1/admin/courses/{courseId}/manage
// Satu respons: data kelas, semua konten (module/article/quiz/zoom/recording/test), paket landing yang memuat kelas, tryout terhubung.
func AdminCourseManageGet(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := chi.URLParam(r, "courseId")
		if courseID == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "courseId required")
			return
		}
		c, err := deps.AdminService.GetCourseByID(r.Context(), courseID)
		if err != nil {
			writeError(w, http.StatusNotFound, "not_found", "course not found")
			return
		}
		contents, err := deps.AdminService.ListCourseContents(r.Context(), courseID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		byType := map[string]int{}
		outContents := make([]dto.CourseContentResponse, 0, len(contents))
		for i := range contents {
			outContents = append(outContents, courseContentToDTO(contents[i]))
			byType[contents[i].Type]++
		}

		resp := dto.AdminCourseManageResponse{
			Course:         adminCourseToDTO(c),
			Contents:       outContents,
			ContentsByType: byType,
			RelatedEndpoints: dto.RelatedCourseAdminEndpoints{
				ListContents:         "/api/v1/admin/courses/" + courseID + "/contents",
				CreateContent:        "/api/v1/admin/courses/" + courseID + "/contents",
				UpdateContent:        "/api/v1/admin/courses/" + courseID + "/contents/{contentId}",
				DeleteContent:        "/api/v1/admin/courses/" + courseID + "/contents/{contentId}",
				ListEnrollments:      "/api/v1/admin/courses/" + courseID + "/enrollments",
				TryoutQuestions:      "/api/v1/admin/tryouts/{tryoutId}/questions",
				PackageManage:        "PUT /api/v1/admin/landing/packages/{id} body linked_course_ids, atau PUT .../courses/{id}/linked-packages",
				GetProgram:           "/api/v1/admin/courses/" + courseID + "/program",
				PutProgram:           "PUT /api/v1/admin/courses/" + courseID + "/program",
				UploadCourseMaterial: "POST /api/v1/admin/upload/course-material (multipart field: file)",
			},
		}

		if deps.CourseAdminLinkRepo != nil {
			pkgs, err := deps.CourseAdminLinkRepo.ListPackagesForCourse(r.Context(), courseID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "server_error", err.Error())
				return
			}
			for _, p := range pkgs {
				resp.LinkedPackages = append(resp.LinkedPackages, dto.AdminCourseLinkedPackage{
					ID: p.ID, Name: p.Name, Slug: p.Slug,
				})
			}
			tryouts, err := deps.CourseAdminLinkRepo.ListTryoutsForCourse(r.Context(), courseID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "server_error", err.Error())
				return
			}
			for _, t := range tryouts {
				resp.LinkedTryouts = append(resp.LinkedTryouts, dto.AdminCourseLinkedTryout{
					ID: t.ID, Title: t.Title, Status: t.Status,
					OpensAt: t.OpensAt, ClosesAt: t.ClosesAt, SortOrder: t.SortOrder,
				})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// AdminCourseLinkedPackagesPut PUT /api/v1/admin/courses/{courseId}/linked-packages
// Mengganti daftar paket landing yang memuat course ini (menulis package_courses).
func AdminCourseLinkedPackagesPut(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.CourseAdminLinkRepo == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "course link repo not configured")
			return
		}
		courseID := chi.URLParam(r, "courseId")
		if courseID == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "courseId required")
			return
		}
		if _, err := deps.AdminService.GetCourseByID(r.Context(), courseID); err != nil {
			writeError(w, http.StatusNotFound, "not_found", "course not found")
			return
		}
		var req dto.AdminCourseLinkedPackagesPutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		if err := deps.CourseAdminLinkRepo.ReplacePackagesForCourse(r.Context(), courseID, req.PackageIDs); err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		cache.InvalidatePackagesList(r.Context(), deps.Redis)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "linked packages updated"})
	}
}

// AdminCourseLinkedTryoutsPut PUT /api/v1/admin/courses/{courseId}/linked-tryouts
// Mengganti daftar tryout yang dihubungkan ke course (course_tryouts).
func AdminCourseLinkedTryoutsPut(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.CourseAdminLinkRepo == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "course link repo not configured")
			return
		}
		courseID := chi.URLParam(r, "courseId")
		if courseID == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "courseId required")
			return
		}
		if _, err := deps.AdminService.GetCourseByID(r.Context(), courseID); err != nil {
			writeError(w, http.StatusNotFound, "not_found", "course not found")
			return
		}
		var req dto.AdminCourseLinkedTryoutsPutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		if err := deps.CourseAdminLinkRepo.ReplaceTryoutsForCourse(r.Context(), courseID, req.TryoutIDs); err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "linked tryouts updated"})
	}
}
