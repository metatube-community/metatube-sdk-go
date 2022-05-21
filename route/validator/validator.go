package validator

type Validator interface {
	Valid(string) bool
}
