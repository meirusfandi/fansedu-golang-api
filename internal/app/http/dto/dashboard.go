package dto

type DashboardSummary struct {
	TotalAttempts int     `json:"totalAttempts"`
	AvgScore      float64 `json:"avgScore"`
	AvgPercentile float64 `json:"avgPercentile"`
}

// DashboardUser data user untuk dashboard siswa
type DashboardUser struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
	SchoolID    *string `json:"schoolId,omitempty"`
	SubjectID   *string `json:"subjectId,omitempty"`
	SchoolName  *string `json:"schoolName,omitempty"`
	SubjectName *string `json:"subjectName,omitempty"`
}

// QuestionScoreDetail satu soal dalam detail penilaian (untuk learning_evaluation)
type QuestionScoreDetail struct {
	QuestionID   string  `json:"questionId"`
	QuestionType string  `json:"questionType"`
	MaxScore     float64 `json:"maxScore"`
	ScoreGot     float64 `json:"scoreGot"`
	Status       string  `json:"status"`
}

// LearningEvaluationDTO detail penilaian attempt + rekomendasi belajar & area yang perlu ditingkatkan
type LearningEvaluationDTO struct {
	AttemptID        string                `json:"attemptId,omitempty"`
	AnswerBreakdown  []QuestionScoreDetail `json:"answerBreakdown"`
	StrengthAreas    []string              `json:"strengthAreas"`
	ImprovementAreas []string              `json:"improvementAreas"`
	Recommendation   string                `json:"recommendation"`
}

type DashboardResponse struct {
	User               DashboardUser          `json:"user"`
	Summary            DashboardSummary       `json:"summary"`
	OpenTryouts        []interface{}          `json:"openTryouts"`
	RecentAttempts     []interface{}          `json:"recentAttempts"`
	CourseProgress     []StudentCourseItem    `json:"courseProgress"`
	StrengthAreas      []string               `json:"strengthAreas"`
	ImprovementAreas   []string               `json:"improvementAreas"`
	Recommendation     string                 `json:"recommendation"`
	LearningEvaluation *LearningEvaluationDTO `json:"learningEvaluation,omitempty"`
}

// GeneralDashboardResponse untuk GET /dashboard (umum, tanpa auth)
type GeneralDashboardResponse struct {
	SiteName      string `json:"siteName"`
	OpenTryouts   int    `json:"openTryouts"`
	TotalCourses  int    `json:"totalCourses"`
	TotalLevels   int    `json:"totalLevels"`
	TotalSubjects int    `json:"totalSubjects"`
	TotalSchools  int    `json:"totalSchools"`
	TotalStudents int    `json:"totalStudents,omitempty"`
}

// --- Student dashboard (new contract) ---

type StudentTryoutSummary struct {
	AttemptedCount  int     `json:"attemptedCount"`
	CompletedCount  int     `json:"completedCount"`
	RegisteredCount int     `json:"registeredCount"`
	AverageScore    float64 `json:"averageScore"`
	BestScore       float64 `json:"bestScore"`
	UpcomingCount   int     `json:"upcomingCount"`
	StreakDays      int     `json:"streakDays"`
	LastAttemptAt   string  `json:"lastAttemptAt"`
}

type StudentWeeklyTarget struct {
	TargetLessons    int `json:"targetLessons"`
	TargetTryouts    int `json:"targetTryouts"`
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
	CoursesCount  int                  `json:"coursesCount"`
	RecentCourses []StudentCourseItem  `json:"recentCourses"`
	TryoutSummary StudentTryoutSummary `json:"tryoutSummary"`
	WeeklyTarget  StudentWeeklyTarget  `json:"weeklyTarget"`
	Badges        []StudentBadge       `json:"badges"`
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
	IsRegistered bool    `json:"isRegistered"`
	HasAttempted bool    `json:"hasAttempted"`
	CanRetake    bool    `json:"canRetake"`
	AttemptCount int     `json:"attemptCount"`
	LastAttemptID *string `json:"lastAttemptId"`
	// Waktu & status tryout (untuk tombol daftar / mulai di frontend)
	OpensAt      string `json:"opensAt"`
	ClosesAt     string `json:"closesAt"`
	TryoutStatus string `json:"tryoutStatus"`
	// CanRegister: belum terdaftar dan masih bisa daftar (tryout open, belum lewat closesAt).
	CanRegister bool `json:"canRegister"`
	// CanStartExam: sudah terdaftar dan boleh buka ujian (resume in_progress atau jendela waktu OK).
	CanStartExam bool `json:"canStartExam"`
	// StartDisabledReason: kode saat CanStartExam false, mis. not_registered, before_opens_at, already_submitted.
	StartDisabledReason string `json:"startDisabledReason,omitempty"`
}

type StudentTryoutHistoryItem struct {
	TryoutID                 string  `json:"tryoutId"`
	TryoutTitle              string  `json:"tryoutTitle"`
	AttemptID                string  `json:"attemptId"`
	Score                    float64 `json:"score"`
	SubmittedAt              string  `json:"submittedAt"`
	ImprovementFromPrevious float64 `json:"improvementFromPrevious"`
}

type StudentTryoutHistoryResponse struct {
	Data []StudentTryoutHistoryItem `json:"data"`
}
