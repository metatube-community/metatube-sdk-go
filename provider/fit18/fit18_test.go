package fit18

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
	"github.com/stretchr/testify/assert"
)

func TestFit18_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"shinaryen",
	})
}

func TestFit18_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"shinaryen-rean:scene1",
	})
}

func TestFit18_GetMovieInfoByURL(t *testing.T) {
	testkit.Test(t, New, []string{
		"https://fit18.com/videos/shinaryen-rean:scene1",
	})
}

func TestThicc18_Name(t *testing.T) {
	provider := NewThicc18()
	assert.Equal(t, "Thicc18", provider.Name())
}

func TestFit18_ParseMovieIDFromURL(t *testing.T) {
	provider := New()

	tests := []struct {
		name     string
		url      string
		expected string
		err      bool
	}{
		{
			name:     "Valid Fit18 URL",
			url:      "https://fit18.com/videos/talent:scene123",
			expected: "talent:scene123",
			err:      false,
		},
		{
			name:     "Valid Fit18 URL with trailing slash",
			url:      "https://fit18.com/videos/talent:scene123/",
			expected: "talent:scene123",
			err:      false,
		},
		{
			name:     "Invalid URL - wrong domain",
			url:      "https://example.com/videos/talent:scene123",
			expected: "",
			err:      true,
		},
		{
			name:     "Invalid URL - wrong path structure",
			url:      "https://fit18.com/talent:scene123",
			expected: "",
			err:      true,
		},
		{
			name:     "Invalid URL - missing colon in ID",
			url:      "https://fit18.com/videos/scene123",
			expected: "",
			err:      true,
		},
		{
			name:     "Invalid URL - malformed",
			url:      "://invalid-url",
			expected: "",
			err:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := provider.ParseMovieIDFromURL(tt.url)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, id)
			}
		})
	}
}

func TestFit18_ExtractNumberFromID(t *testing.T) {
	provider := New()

	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "Standard ID format",
			id:       "talent:scene123",
			expected: "scene123",
		},
		{
			name:     "ID without colon",
			id:       "scene123",
			expected: "scene123",
		},
		{
			name:     "Multiple colons",
			id:       "talent:sub:scene123",
			expected: "scene123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.extractNumberFromID(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}
