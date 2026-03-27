package dto

// --- Programs (katalog) ---
type ProgramsListResponse struct {
	Data       []ProgramListItem `json:"data"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	TotalPages int               `json:"totalPages"`
}

type ProgramListItem struct {
	ID               string           `json:"id"`
	Slug             string           `json:"slug"`
	Title            string           `json:"title"`
	ShortDescription string           `json:"shortDescription"`
	Thumbnail        string           `json:"thumbnail"`
	Price            int              `json:"price"`
	PriceDisplay     string           `json:"priceDisplay"`
	Guru             ProgramGuru `json:"guru"`
	Category         string           `json:"category"`
	Level            string           `json:"level"`
	Duration         string           `json:"duration"`
	Rating           float64          `json:"rating"`
	ReviewCount      int              `json:"reviewCount"`
}

type ProgramGuru struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type ProgramDetailResponse struct {
	ID               string            `json:"id"`
	Slug             string            `json:"slug"`
	Title            string            `json:"title"`
	ShortDescription string            `json:"shortDescription"`
	Description      string            `json:"description"`
	Thumbnail        string            `json:"thumbnail"`
	Price            int               `json:"price"`
	PriceDisplay     string            `json:"priceDisplay"`
	Guru             ProgramGuru `json:"guru"`
	Category         string            `json:"category"`
	Level            string            `json:"level"`
	Duration         string            `json:"duration"`
	Rating           float64            `json:"rating"`
	ReviewCount      int                `json:"reviewCount"`
	Modules         []ProgramModule    `json:"modules"`
	Reviews         []ProgramReview    `json:"reviews"`
}

type ProgramModule struct {
	ID      string           `json:"id"`
	Title   string           `json:"title"`
	Lessons []ProgramLesson  `json:"lessons"`
}

type ProgramLesson struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Duration string `json:"duration"`
}

type ProgramReview struct {
	ID      string `json:"id"`
	User    string `json:"user"`
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
	Date    string `json:"date"`
}

// --- Student courses (enrolled with progress) ---
type StudentCoursesResponse struct {
	Data       []StudentCourseItem `json:"data"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	TotalPages int                  `json:"totalPages"`
}

type StudentCourseItem struct {
	ID              string              `json:"id"`
	Program         StudentCourseProgram `json:"program"`
	ProgressPercent int                 `json:"progressPercent"`
	EnrolledAt      string              `json:"enrolledAt"`
	LastAccessedAt string              `json:"lastAccessedAt"`
}

type StudentCourseProgram struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
}

// --- Student transactions ---
type StudentTransactionsResponse struct {
	Data       []StudentTransactionItem `json:"data"`
	Total      int                    `json:"total"`
	Page       int                    `json:"page"`
	TotalPages int                    `json:"totalPages"`
}

type StudentTransactionItem struct {
	ID                string                     `json:"id"`
	OrderID           string                     `json:"orderId"`
	Status            string                     `json:"status"`
	IsCollective      bool                       `json:"isCollective"`
	Quantity          int                        `json:"quantity"`
	UnitPrice         int                        `json:"unitPrice"`
	Subtotal          int                        `json:"subtotal"`
	UniqueCode        int                        `json:"uniqueCode"`
	Total             int                        `json:"total"`
	NormalPrice       int                        `json:"normalPrice"`
	PromoCode         string                     `json:"promoCode"`
	Discount          int                        `json:"discount"`
	DiscountPercent   float64                    `json:"discountPercent"`
	ConfirmationCode  string                     `json:"confirmationCode"`
	Programs          []StudentTransactionProgram `json:"programs"`
	Students          []CheckoutStudentItem      `json:"students,omitempty"`
	PaidAt            string                     `json:"paidAt"`
}

type StudentTransactionProgram struct {
	Title string `json:"title"`
}

// --- Guru (dashboard LMS) ---
type GuruStudentsResponse struct {
	Data []GuruStudentItem `json:"data"`
}

type GuruStudentItem struct {
	UserID          string `json:"userId"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	ProgramTitle    string `json:"programTitle"`
	ProgressPercent int    `json:"progressPercent"`
}

type GuruEarningsResponse struct {
	Data []GuruEarningItem `json:"data"`
}

type GuruEarningItem struct {
	Period      string `json:"period"`
	Revenue     int64  `json:"revenue"`
	NewStudents int    `json:"newStudents"`
}

// --- Checkout (spec alignment) ---
type CheckoutInitiateRequestLMS struct {
	ProgramID   string `json:"programId"`
	ProgramSlug string `json:"programSlug"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PromoCode   string `json:"promoCode"`
	Quantity    int    `json:"quantity,omitempty"`
	BuyerRole   string `json:"buyerRole,omitempty"`
	Students    []CheckoutStudentItem `json:"students,omitempty"`
}

type CheckoutStudentItem struct {
	Name   string  `json:"name,omitempty"`
	Email  string  `json:"email,omitempty"`
	UserID *string `json:"userId,omitempty"`
}

type CheckoutInitiateResponseLMS struct {
	CheckoutID       string               `json:"checkoutId"`
	OrderID          string               `json:"orderId"`
	Total            int                  `json:"total"`
	Program          CheckoutProgramInfo  `json:"program"`
	NormalPrice      int                  `json:"normalPrice"`
	PromoCode        string               `json:"promoCode,omitempty"`
	Discount          int                  `json:"discount"`
	DiscountPercent  float64              `json:"discountPercent"`
	FinalPrice       int                  `json:"finalPrice"`
	ConfirmationCode string               `json:"confirmationCode"`
	IsCollective     bool                 `json:"isCollective"`
	Quantity         int                  `json:"quantity"`
	UnitPrice        int                  `json:"unitPrice"`
	Subtotal         int                  `json:"subtotal"`
	UniqueCode       int                  `json:"uniqueCode"`
	Students         []CheckoutStudentItem `json:"students,omitempty"`
}

type CheckoutProgramInfo struct {
	Title        string `json:"title"`
	PriceDisplay string `json:"priceDisplay"`
}

type CheckoutPaymentSessionRequestLMS struct {
	CheckoutID     string `json:"checkoutId"`
	PaymentMethod  string `json:"paymentMethod"`
	PromoCode      string `json:"promoCode"`
}

type CheckoutPaymentSessionResponseLMS struct {
	PaymentURL           string `json:"paymentUrl"`
	OrderID              string `json:"orderId"`
	Expiry               string `json:"expiry,omitempty"`
	VirtualAccountNumber string `json:"virtualAccountNumber,omitempty"`
	Amount               int    `json:"amount"`
}

// --- Error (spec) ---
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
