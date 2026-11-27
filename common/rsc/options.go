package rsc

// Option configures FindObject.
type Option func(*options)

type options struct {
	keys []string
}

// RequireKeys only matches objects that have all the keys, case-sensitively.
func RequireKeys(keys ...string) Option {
	return func(o *options) { o.keys = append(o.keys, keys...) }
}
