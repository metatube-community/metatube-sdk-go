package internal

type ResponseWrapper struct {
	BackendResponse BackendResponse `json:"backendResponse"`
}

type BackendResponse struct {
	CategoryLinks []any      `json:"category_links"`
	Contents      Contents   `json:"contents"`
	ExchangeURL   string     `json:"exchange_url"`
	Limit         int        `json:"limit"`
	Navigation    Navigation `json:"navigation"`
	OrderBy       string     `json:"order_by"`
	Page          int        `json:"page"`
	Pagination    Pagination `json:"pagination"`
	SearchWord    string     `json:"search_word"`
}

type Contents struct {
	Count            int       `json:"count"`
	ExceededHitLimit bool      `json:"exceeded_hit_limit"`
	Data             []Product `json:"data"`
}

type Product struct {
	Authors             []string `json:"authors"`
	Casts               []string `json:"casts"`
	Comment             string   `json:"comment"`
	ContentID           string   `json:"content_id"`
	CurrentPrice        int      `json:"current_price"`
	DetailURL           string   `json:"detail_url"`
	DiscountRate        int      `json:"discount_rate"`
	Floor               string   `json:"floor"`
	IsAvailable         bool     `json:"is_available"`
	ItemPricing         string   `json:"item_pricing"`
	Keywords            []string `json:"keywords"`
	Makers              []string `json:"makers"`
	MediaType           string   `json:"media_type"`
	OriginalPrice       int      `json:"original_price"`
	Rate                float64  `json:"rate"`
	ReleaseAnnouncement string   `json:"release_announcement"`
	ReleaseStatuses     []string `json:"release_statuses"`
	ReviewCount         int      `json:"review_count"`
	SaleStatus          string   `json:"sale_status"`
	SampleMovieURL      string   `json:"sample_movie_url"`
	SearchServiceURL    string   `json:"search_service_url"`
	Series              string   `json:"series"`
	ServiceName         string   `json:"service_name"`
	ThumbnailImageURL   string   `json:"thumbnail_image_url"`
	Title               string   `json:"title"`
}

type Navigation struct {
	List []NavItem `json:"list"`
}

type NavItem struct {
	Service string `json:"service"`
	URL     string `json:"url"`
	Count   int    `json:"count"`
}

type Pagination struct {
	CurrentPage int `json:"currentPage"`
	PagesCount  int `json:"pagesCount"`
}
