package avbase

type workResponse struct {
	ID      int    `json:"id"`
	Prefix  string `json:"prefix"`
	WorkID  string `json:"work_id"`
	Title   string `json:"title"`
	MinDate string `json:"min_date"`
	Genres  []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	Products []struct {
		ID           int    `json:"id"`
		ProductID    string `json:"product_id"`
		URL          string `json:"url"`
		Title        string `json:"title"`
		Source       string `json:"source"`
		ImageURL     string `json:"image_url"`
		ThumbnailURL string `json:"thumbnail_url"`
		Date         string `json:"date"`
		Maker        struct {
			Name string `json:"name"`
		} `json:"maker"`
		Label struct {
			Name string `json:"name"`
		} `json:"label"`
		Series struct {
			Name string `json:"name"`
		} `json:"series"`
		SampleImageURLS []struct {
			S string `json:"s"`
			L string `json:"l"`
		} `json:"sample_image_urls"`
		ItemInfo struct {
			Description string `json:"description"`
			Price       string `json:"price"`
			Volume      string `json:"volume"`
		} `json:"iteminfo"`
	} `json:"products"`
	Actors []actorResponse `json:"actors"`
	Casts  []struct {
		Actor actorResponse `json:"actor"`
	} `json:"casts"`
}

type actorResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}
