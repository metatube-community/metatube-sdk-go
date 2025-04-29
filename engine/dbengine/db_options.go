package dbengine

type ActorSearchOptions struct {
	Provider  string
	Threshold float64
	Limit     int
	Offset    int
}

func (opts *ActorSearchOptions) applyDefaults() {
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

type MovieSearchOptions struct {
	Provider   string
	Thresholds MovieThresholds
	Limit      int
	Offset     int
}

type MovieThresholds struct {
	Number float64
	Title  float64
}

func (opts *MovieSearchOptions) applyDefaults() {
	const (
		maxLimit        = 20
		numberThreshold = 0.4
		titleThreshold  = 0.2
	)
	if opts.Thresholds.Number == 0 {
		opts.Thresholds.Number = numberThreshold
	}
	if opts.Thresholds.Title == 0 {
		opts.Thresholds.Title = titleThreshold
	}
	if opts.Limit > maxLimit {
		opts.Limit = maxLimit
	}
}
