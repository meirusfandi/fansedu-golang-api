package domain

// PackageLinkedCourse is an LMS course attached to a landing package (for bundles / multi-class access).
type PackageLinkedCourse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug,omitempty"`
}

// LandingPackage is a package/program for the landing page (snake_case in API response).
type LandingPackage struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Slug                 string   `json:"slug"`
	ShortDescription     *string  `json:"short_description,omitempty"`
	PriceEarlyBird  *int64   `json:"price_early_bird,omitempty"`  // nominal dalam rupiah
	PriceNormal     *int64   `json:"price_normal,omitempty"`     // nominal dalam rupiah
	CTALabel             string   `json:"cta_label"`
	WAMessageTemplate  *string  `json:"wa_message_template,omitempty"`
	CTAURL             *string  `json:"cta_url,omitempty"`
	IsOpen             bool     `json:"is_open"`
	IsBundle           bool     `json:"is_bundle"`
	BundleSubtitle     *string  `json:"bundle_subtitle,omitempty"`
	Durasi             *string  `json:"durasi,omitempty"`
	Materi             []string `json:"materi,omitempty"`
	Fasilitas          []string `json:"fasilitas,omitempty"`
	Bonus              []string `json:"bonus,omitempty"`
	LinkedCourses      []PackageLinkedCourse `json:"linked_courses,omitempty"`
}

// LandingPackagePriceRupiah returns display/checkout price: early bird if set, else normal.
func LandingPackagePriceRupiah(pkg LandingPackage) int {
	if pkg.PriceEarlyBird != nil && *pkg.PriceEarlyBird > 0 {
		return int(*pkg.PriceEarlyBird)
	}
	if pkg.PriceNormal != nil && *pkg.PriceNormal > 0 {
		return int(*pkg.PriceNormal)
	}
	return 0
}
