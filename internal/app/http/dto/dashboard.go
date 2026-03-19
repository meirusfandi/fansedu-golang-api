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

// QuestionScoreDetail satu soal dalam detail penilaian (untuk learning_evaluation)
type QuestionScoreDetail struct {
	QuestionID   string  `json:"question_id"`
	QuestionType string  `json:"question_type"`
	MaxScore     float64 `json:"max_score"`
	ScoreGot     float64 `json:"score_got"`
	Status       string  `json:"status"`
}

// LearningEvaluationDTO detail penilaian attempt + rekomendasi belajar & area yang perlu ditingkatkan
type LearningEvaluationDTO struct {
	AttemptID       string                `json:"attempt_id,omitempty"`
	AnswerBreakdown []QuestionScoreDetail `json:"answer_breakdown"`
	StrengthAreas   []string              `json:"strength_areas"`
	ImprovementAreas []string             `json:"improvement_areas"`
	Recommendation  string                `json:"recommendation"`
}

type DashboardResponse struct {
	User                DashboardUser        `json:"user"`
	Summary             DashboardSummary    `json:"summary"`
	OpenTryouts         []interface{}       `json:"open_tryouts"`
	RecentAttempts      []interface{}       `json:"recent_attempts"`
	CourseProgress     []StudentCourseItem `json:"course_progress"`
	StrengthAreas       []string            `json:"strength_areas"`
	ImprovementAreas    []string            `json:"improvement_areas"`
	Recommendation      string              `json:"recommendation"`
	LearningEvaluation  *LearningEvaluationDTO `json:"learning_evaluation,omitempty"`
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

// --- Student dashboard (new contract) ---

type StudentTryoutSummary struct {
	AttemptedCount int     `json:"attemptedCount"`
	CompletedCount int     `json:"completedCount"`
	RegisteredCount int    `json:"registeredCount"`
	AverageScore   float64 `json:"averageScore"`
	BestScore      float64 `json:"bestScore"`
	UpcomingCount  int      `json:"upcomingCount"`
	StreakDays     int      `json:"streakDays"`
	LastAttemptAt  string   `json:"lastAttemptAt"`
}

type StudentWeeklyTarget struct {
	TargetLessons   int `json:"targetLessons"`
	TargetTryouts   int `json:"targetTryouts"`
	CompletedLessons int `json:"completedLessons"`
	CompletedTryouts int `json:"completedTryouts"`
}

type StudentBadge struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	EarnedAt    string `json:"earnedAt"`
}

type StudentDashboardResponse struct {
	CoursesCount   int                   `json:"coursesCount"`
	RecentCourses  []StudentCourseItem `json:"recentCourses"`
	TryoutSummary  StudentTryoutSummary `json:"tryoutSummary"`
	WeeklyTarget   StudentWeeklyTarget  `json:"weeklyTarget"`
	Badges         []StudentBadge       `json:"badges"`
}

// --- Next actions ---

type StudentNextAction struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Href        string `json:"href"`
	Priority    int    `json:"priority"`
}

type StudentNextActionsResponse struct {
	Data []StudentNextAction `json:"data"`
}

// --- Tryout status & history ---

type StudentTryoutStatusResponse struct {
	IsRegistered  bool    `json:"isRegistered"`
	HasAttempted  bool    `json:"hasAttempted"`
	CanRetake     bool    `json:"canRetake"`
	AttemptCount  int     `json:"attemptCount"`
	LastAttemptID *string `json:"lastAttemptId"`
}

type StudentTryoutHistoryItem struct {
	TryoutID                  string  `json:"tryoutId"`
	TryoutTitle               string  `json:"tryoutTitle"`
	AttemptID                 string  `json:"attemptId"`
	Score                     float64 `json:"score"`
	SubmittedAt               string  `json:"submittedAt"`
	ImprovementFromPrevious  float64 `json:"improvementFromPrevious"`
}

type StudentTryoutHistoryResponse struct {
	Data []StudentTryoutHistoryItem `json:"data"`
}
