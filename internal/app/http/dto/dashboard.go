package dto

type DashboardSummary struct {
	TotalAttempts  int     `json:"total_attempts"`
	AvgScore       float64 `json:"avg_score"`
	AvgPercentile  float64 `json:"avg_percentile"`
}

// DashboardUser data user untuk dashboard siswa
type DashboardUser struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	SchoolID    *string `json:"school_id,omitempty"`
	SubjectID   *string `json:"subject_id,omitempty"`
	SchoolName  *string `json:"school_name,omitempty"`
	SubjectName *string `json:"subject_name,omitempty"`
}

type DashboardResponse struct {
	User              DashboardUser   `json:"user"`
	Summary           DashboardSummary `json:"summary"`
	OpenTryouts       []interface{}   `json:"open_tryouts"`
	RecentAttempts    []interface{}  `json:"recent_attempts"`
	StrengthAreas     []string        `json:"strength_areas"`
	ImprovementAreas  []string        `json:"improvement_areas"`
	Recommendation    string          `json:"recommendation"`
}

// GeneralDashboardResponse untuk GET /dashboard (umum, tanpa auth)
type GeneralDashboardResponse struct {
	SiteName        string `json:"site_name"`
	OpenTryouts     int    `json:"open_tryouts"`
	TotalCourses    int    `json:"total_courses"`
	TotalLevels     int    `json:"total_levels"`
	TotalSubjects   int    `json:"total_subjects"`
	TotalSchools    int    `json:"total_schools"`
	TotalStudents   int    `json:"total_students,omitempty"`
}
