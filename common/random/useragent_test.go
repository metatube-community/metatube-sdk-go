package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserAgent(t *testing.T) {
	// We should have at least 100 User-Agent(s).
	assert.Greater(t, len(_userAgents), 100)
	for _, ua := range _userAgents {
		assert.NotEmpty(t, ua.Raw)
		assert.True(t, filter(ua))
	}
	for i := 0; i < 1e2; i++ {
		assert.NotEmpty(t, UserAgent())
	}
}
