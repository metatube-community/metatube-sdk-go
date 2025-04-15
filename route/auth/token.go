package auth

var (
	_ Validator = (*Token)(nil)
	_ Validator = (*TokenStore)(nil)
)

type Token string

func (token Token) Valid(t string) bool {
	return string(token) == t
}

type TokenStore map[string]struct{}

func NewTokenStore(tokens ...string) TokenStore {
	store := make(TokenStore)
	store.Add(tokens...)
	return store
}

func (store TokenStore) Add(tokens ...string) {
	for _, token := range tokens {
		store[token] = struct{}{}
	}
}

func (store TokenStore) Del(token string) {
	delete(store, token)
}

func (store TokenStore) Valid(token string) (ok bool) {
	_, ok = store[token]
	return
}
