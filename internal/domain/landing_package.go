package domain

// PackageLinkedCourse is an LMS course attached to a landing package (for bundles / multi-class access).
type PackageLinkedCourse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug,omitempty"`
}

// LandingPackage is a package/program for the landing page (JSON camelCase).
type LandingPackage struct {
	ID                string                `json:"id"`
	Name              string                `json:"name"`
	Slug              string                `json:"slug"`
	ShortDescription  *string               `json:"shortDescription,omitempty"`
	PriceEarlyBird    *int64                `json:"priceEarlyBird,omitempty"`
	PriceNormal       *int64                `json:"priceNormal,omitempty"`
	CTALabel          string                `json:"ctaLabel"`
	WAMessageTemplate *string               `json:"waMessageTemplate,omitempty"`
	CTAURL            *string               `json:"ctaUrl,omitempty"`
	IsOpen            bool                  `json:"isOpen"`
	IsBundle          bool                  `json:"isBundle"`
	BundleSubtitle    *string               `json:"bundleSubtitle,omitempty"`
	Durasi            *string               `json:"durasi,omitempty"`
	Materi            []string              `json:"materi,omitempty"`
	Fasilitas         []string              `json:"fasilitas,omitempty"`
	Bonus             []string              `json:"bonus,omitempty"`
	LinkedCourses     []PackageLinkedCourse `json:"linkedCourses,omitempty"`
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
