package graphql

type SearchVideosOptions struct {
	Query string
	Site  string
	First int
	Skip  int
}

type FindOneVideoOptions struct {
	Slug string
	Site string
}

type UserReviewsOptions struct {
	Slug string
	Site string
}
