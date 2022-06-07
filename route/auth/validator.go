package auth

type Validator interface {
	Valid(string) bool
}
