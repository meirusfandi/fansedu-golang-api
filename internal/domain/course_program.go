package domain

// CourseProgramMeeting satu pertemuan (1–8) untuk admin — disinkron ke course_sections + learning_lessons.
type CourseProgramMeeting struct {
	MeetingNumber   int
	Title           string
	DetailText      *string
	PdfURL          *string
	PrTitle         *string
	PrDescription   *string
	LiveClassURL    *string
}
