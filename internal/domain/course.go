package domain

import "time"

type Course struct {
	ID          string
	Title       string
	Slug        *string
	Description *string
	Price       int // nominal dalam rupiah
	Thumbnail   *string
	SubjectID   *string
	CreatedBy   *string
	// TrackType: "meetings" = alur pertemuan (PDF/PR/live); "tryout" = fokus tryout terhubung.
	TrackType   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const CourseTrackMeetings = "meetings"
const CourseTrackTryout   = "tryout"
