package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
)

func registerClassOptions(levelSlug string) []dto.RegisterClassOption {
	switch strings.ToLower(strings.TrimSpace(levelSlug)) {
	case "sd":
		return []dto.RegisterClassOption{
			{Value: "1", Label: "Kelas 1 SD"},
			{Value: "2", Label: "Kelas 2 SD"},
			{Value: "3", Label: "Kelas 3 SD"},
			{Value: "4", Label: "Kelas 4 SD"},
			{Value: "5", Label: "Kelas 5 SD"},
			{Value: "6", Label: "Kelas 6 SD"},
		}
	case "smp":
		return []dto.RegisterClassOption{
			{Value: "7", Label: "Kelas 7 SMP"},
			{Value: "8", Label: "Kelas 8 SMP"},
			{Value: "9", Label: "Kelas 9 SMP"},
		}
	case "sma":
		return []dto.RegisterClassOption{
			{Value: "10", Label: "Kelas 10 SMA"},
			{Value: "11", Label: "Kelas 11 SMA"},
			{Value: "12", Label: "Kelas 12 SMA"},
		}
	default:
		return []dto.RegisterClassOption{}
	}
}

// AuthRegisterMasterData GET /api/v1/auth/register/master-data
// Mengembalikan jenjang pendidikan + bidang pelajaran + opsi kelas per jenjang.
func AuthRegisterMasterData(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps == nil || deps.LevelRepo == nil || deps.SubjectRepo == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "master data unavailable")
			return
		}
		levels, err := deps.LevelRepo.List(r.Context())
		if err != nil {
			writeInternalError(w, r, err)
			return
		}

		out := make([]dto.RegisterLevelOption, 0, len(levels))
		for _, lv := range levels {
			subjectIDs, err := deps.LevelRepo.ListSubjectIDsByLevel(r.Context(), lv.ID)
			if err != nil {
				writeInternalError(w, r, err)
				return
			}

			subjects := make([]dto.RegisterSubjectOption, 0, len(subjectIDs))
			for _, sid := range subjectIDs {
				s, err := deps.SubjectRepo.GetByID(r.Context(), sid)
				if err != nil {
					continue
				}
				subjects = append(subjects, dto.RegisterSubjectOption{
					ID:   s.ID,
					Name: s.Name,
					Slug: s.Slug,
				})
			}

			out = append(out, dto.RegisterLevelOption{
				ID:          lv.ID,
				Name:        lv.Name,
				Slug:        lv.Slug,
				Description: lv.Description,
				Classes:     registerClassOptions(lv.Slug),
				Subjects:    subjects,
			})
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(dto.RegisterMasterDataResponse{
			Levels: out,
		})
	}
}

