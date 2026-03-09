package domain

// LandingPackage is a package/program for the landing page (snake_case in API response).
type LandingPackage struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Slug               string   `json:"slug"`
	ShortDescription   *string  `json:"short_description,omitempty"`
	PriceDisplay       *string  `json:"price_display,omitempty"`
	PriceEarlyBird     *string  `json:"price_early_bird,omitempty"`
	PriceNormal        *string  `json:"price_normal,omitempty"`
	CTALabel           string   `json:"cta_label"`
	WAMessageTemplate  *string  `json:"wa_message_template,omitempty"`
	CTAURL             *string  `json:"cta_url,omitempty"`
	IsOpen             bool     `json:"is_open"`
	IsBundle           bool     `json:"is_bundle"`
	BundleSubtitle     *string  `json:"bundle_subtitle,omitempty"`
	Durasi             *string  `json:"durasi,omitempty"`
	Materi             []string `json:"materi,omitempty"`
	Fasilitas          []string `json:"fasilitas,omitempty"`
	Bonus              []string `json:"bonus,omitempty"`
}
