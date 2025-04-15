package providerid

import (
	"encoding"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	_ fmt.Stringer             = (*ProviderID)(nil)
	_ encoding.TextMarshaler   = (*ProviderID)(nil)
	_ encoding.TextUnmarshaler = (*ProviderID)(nil)
)

var ErrInvalidProviderID = errors.New("invalid provider/id pair")

type ProviderID struct {
	Provider, ID string
}

func New(provider, id string) (ProviderID, error) {
	pid := ProviderID{
		Provider: provider,
		ID:       id,
	}
	if !pid.IsValid() {
		return ProviderID{}, ErrInvalidProviderID
	}
	return pid, nil
}

func Parse(s string) (ProviderID, error) {
	const separator = ":"
	provider, id, found := strings.Cut(trimExtraArgs(s), separator)
	if !found {
		return ProviderID{}, ErrInvalidProviderID
	}
	// unescape id part.
	id, err := url.QueryUnescape(id)
	if err != nil {
		return ProviderID{}, ErrInvalidProviderID
	}
	return New(provider, id)
}

func trimExtraArgs(s string) string {
	return regexp.MustCompile(`:[0|1](\.\d+)?$`).ReplaceAllString(s, "")
}

func (pid *ProviderID) IsValid() bool {
	return pid.Provider != "" && pid.ID != ""
}

func (pid *ProviderID) String() string {
	return fmt.Sprintf("%s:%s", pid.Provider, url.QueryEscape(pid.ID))
}

func (pid *ProviderID) MarshalText() (text []byte, err error) {
	return []byte(pid.String()), nil
}

func (pid *ProviderID) UnmarshalText(text []byte) error {
	p, err := Parse(string(text))
	if err != nil {
		return err
	}
	pid.Provider = p.Provider
	pid.ID = p.ID
	return nil
}
