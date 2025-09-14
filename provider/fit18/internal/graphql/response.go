package graphql

// SearchResponse represents the response from the search API
type SearchResponse struct {
	Search struct {
		Search struct {
			Result []SearchResultItem `json:"result"`
		} `json:"search"`
	} `json:"search"`
}

// SearchResultItem represents a single search result item
type SearchResultItem struct {
	Type        string   `json:"type"`
	ItemID      string   `json:"itemId"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Images      []string `json:"images"`
	Typename    string   `json:"__typename"`
}

// BatchFindAssetResponse represents the response from the batch asset API
type BatchFindAssetResponse struct {
	Asset struct {
		Batch struct {
			Result []AssetResult `json:"result"`
		} `json:"batch"`
	} `json:"asset"`
}

// AssetResult represents a single asset result
type AssetResult struct {
	Path  string `json:"path"`
	Mime  string `json:"mime"`
	Size  int    `json:"size"`
	Serve struct {
		Type     string `json:"type"`
		URI      string `json:"uri"`
		Typename string `json:"__typename"`
	} `json:"serve"`
	Typename string `json:"__typename"`
}

// FindVideoResponse represents the response from the find video API
type FindVideoResponse struct {
	Data struct {
		Video struct {
			Find struct {
				Result struct {
					VideoID      string `json:"videoId"`
					Title        string `json:"title"`
					Duration     int    `json:"duration"`
					GalleryCount int    `json:"galleryCount"`
					Description  struct {
						Short    string `json:"short"`
						Long     string `json:"long"`
						Typename string `json:"__typename"`
					} `json:"description"`
					Talent []struct {
						Type   string `json:"type"`
						Talent struct {
							TalentID string `json:"talentId"`
							Name     string `json:"name"`
							Typename string `json:"__typename"`
						} `json:"talent"`
						Typename string `json:"__typename"`
					} `json:"talent"`
					Typename string `json:"__typename"`
				} `json:"result"`
				Typename string `json:"__typename"`
			} `json:"find"`
			Typename string `json:"__typename"`
		} `json:"video"`
	} `json:"data"`
}
