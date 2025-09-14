package graphql

type SearchVideosResponse struct {
	SearchVideos SearchVideosData `json:"searchVideos"`
}

type SearchVideosData struct {
	Edges []SearchVideoEdge `json:"edges"`
}

type SearchVideoEdge struct {
	Node SearchVideoNode `json:"node"`
}

type SearchVideoNode struct {
	VideoId     string           `json:"videoId"`
	Title       string           `json:"title"`
	ReleaseDate string           `json:"releaseDate"`
	Slug        string           `json:"slug"`
	Images      SearchVideoImage `json:"images"`
}

type SearchVideoImage struct {
	Listing []struct {
		Src string `json:"src"`
	} `json:"listing"`
}

type SearchResultsTourResponse struct {
	SearchCategories []SearchCategory `json:"searchCategories"`
	SearchModels     []SearchModel    `json:"searchModels"`
	SearchVideos     SearchVideosTour `json:"searchVideos"`
}

type SearchCategory struct {
	CategoryId string `json:"categoryId"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	Typename   string `json:"__typename"`
}

type SearchModel struct {
	Name   string      `json:"name"`
	Slug   string      `json:"slug"`
	Images ModelImages `json:"images"`
	Typename string    `json:"__typename"`
}

type ModelImages struct {
	Listing []ImageInfo `json:"listing"`
	Typename string     `json:"__typename"`
}

type SearchVideosTour struct {
	Edges    []SearchVideoTourEdge `json:"edges"`
	Typename string                `json:"__typename"`
}

type SearchVideoTourEdge struct {
	Node     SearchVideoTourNode `json:"node"`
	Typename string              `json:"__typename"`
}

type SearchVideoTourNode struct {
	Description string            `json:"description"`
	Title       string            `json:"title"`
	Slug        string            `json:"slug"`
	Models      []ModelSlugged    `json:"modelsSlugged"`
	Images      VideoTourImages   `json:"images"`
	Typename    string            `json:"__typename"`
}

type ModelSlugged struct {
	Name     string `json:"name"`
	Slug     string `json:"slugged"`
	Typename string `json:"__typename"`
}

type VideoTourImages struct {
	Listing  []ImageInfo `json:"listing"`
	Typename string      `json:"__typename"`
}

type ImageInfo struct {
	Src        string     `json:"src"`
	Placeholder string     `json:"placeholder"`
	Width      int        `json:"width"`
	Height     int        `json:"height"`
	Highdpi    HighDpi    `json:"highdpi"`
	Webp       WebpImages `json:"webp"`
}

type HighDpi struct {
	Double   string `json:"double"`
	Typename string `json:"__typename"`
}

type WebpImages struct {
	Src        string        `json:"src"`
	Placeholder string        `json:"placeholder"`
	Highdpi    HighDpiWebp   `json:"highdpi"`
	Typename   string        `json:"__typename"`
}

type HighDpiWebp struct {
	Double   string `json:"double"`
	Typename string `json:"__typename"`
}

type FindOneVideoResponse struct {
	FindOneVideo FindOneVideoData `json:"findOneVideo"`
}

type FindOneVideoData struct {
	VideoId     string              `json:"videoId"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	ReleaseDate string              `json:"releaseDate"`
	Models      []VideoModel        `json:"models"`
	Directors   []VideoDirector     `json:"directors"`
	Categories  []VideoCategory     `json:"categories"`
	Carousel    []VideoCarouselItem `json:"carousel"`
	Reviews     VideoReviews        `json:"reviews"`
}

type VideoModel struct {
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Images struct {
		Listing []struct {
			Highdpi struct {
				Double string `json:"double"`
			} `json:"highdpi"`
		} `json:"listing"`
	} `json:"images"`
}

type VideoDirector struct {
	Name string `json:"name"`
}

type VideoCategory struct {
	Name string `json:"name"`
}

type VideoCarouselItem struct {
	Listing []struct {
		Highdpi struct {
			Triple string `json:"triple"`
		} `json:"highdpi"`
	} `json:"listing"`
}

type VideoReviews struct {
	Items []UserReviewItem `json:"items"`
}

type UserReviewItem struct {
	Title        string  `json:"title"`
	Rating       float64 `json:"rating"`
	ReviewerId   string  `json:"reviewerId"`
	Nickname     string  `json:"nickname"`
	IsPurchased  bool    `json:"isPurchased"`
	Comment      string  `json:"comment"`
	HelpfulCount int     `json:"helpfulCount"`
	Service      string  `json:"service"`
	IsExposure   bool    `json:"isExposure"`
	PublishDate  string  `json:"publishDate"`
}