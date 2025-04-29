package dbengine

type SearchOptions struct {
	Provider  string
	Threshold float64
	Limit     int
	Offset    int
}

func (opts *SearchOptions) applyDefaults() {
	const (
		maxLimit  = 20
		threshold = 0.2 // be more tolerated for name search.
	)
	if opts.Threshold == 0 {
		opts.Threshold = threshold
	}
	if opts.Limit > maxLimit {
		opts.Limit = maxLimit
	}
}
