package dto

type DashboardSummary struct {
	TotalAttempts  int     `json:"total_attempts"`
	AvgScore       float64 `json:"avg_score"`
	AvgPercentile  float64 `json:"avg_percentile"`
}

type DashboardResponse struct {
	Summary          DashboardSummary `json:"summary"`
	OpenTryouts      []interface{}    `json:"open_tryouts"`
	RecentAttempts   []interface{}    `json:"recent_attempts"`
	StrengthAreas    []string         `json:"strength_areas"`
	ImprovementAreas []string         `json:"improvement_areas"`
	Recommendation   string           `json:"recommendation"`
}
