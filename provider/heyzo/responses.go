package heyzo

// movieResponse is the response of the movie API.
type movieResponse struct {
	MovieSeq         int    `json:"movie_seq"`
	MovieID          string `json:"movie_id"`
	Name             string `json:"name"`
	MonthlyStartDate string `json:"monthly_start_date"`
	IsBonus          bool   `json:"isBonus"`
	IsLimited        bool   `json:"isLimited"`
	Detail           struct {
		Memo    string  `json:"memo"`
		RvScore float64 `json:"rv_score"`
	} `json:"detail"`
	Actors []struct {
		ActorName string `json:"actor_name"`
	} `json:"actors"`
	Categories []struct {
		CategoryName string `json:"category_name"`
		Level        int    `json:"LEVEL"`
	} `json:"categories"`
	Attributes []struct {
		AttributeID    int    `json:"attribute_id"`
		AttributeValue string `json:"attribute_value"`
	} `json:"attributes"`
	Download struct {
		Whole []struct {
			Duration string `json:"duration"`
		} `json:"whole"`
	} `json:"download"`
}

// movieDuration is the JSON encoded value of the duration attribute.
type movieDuration struct {
	Full string `json:"full"`
}

// reviewResponse is the response of the review API.
type reviewResponse struct {
	Items []struct {
		Username string  `json:"username"`
		Score    float64 `json:"score"`
		Comment  string  `json:"comment"`
		Created  string  `json:"created"`
	} `json:"items"`
}
