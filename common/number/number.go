package number

import (
	"path"
	"regexp"
	"strings"
	"unicode"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
)

func findFirstNonASCII(s string) int {
	for i, r := range []rune(s) {
		if r > unicode.MaxASCII {
			return i
		}
	}
	return len(s)
}

func Trim(s string) string {
	const maxExtLength = 7
	if ext := path.Ext(s); len(ext) < maxExtLength {
		s = s[:len(s)-len(ext)] // trim extension
	}
	if ss := strings.SplitN(s, "@", 2); len(ss) == 2 {
		s = ss[1] // trim @ char
	}
	s = regexp.MustCompile(`(?i)\s+-\s+`).
		ReplaceAllString(s, " " /* space */) // fix `number - title` style
	s = regexp.MustCompile(`(?i)[-_](\d*fps|fhd|hd|sd|1080p|720p|uncensored|leak|[2468]K|[xh]26[45])+|\[.*]`).
		ReplaceAllString(s, "") // trim tags
	s = regexp.MustCompile(`^(?i)\s*(cari|carib|caribean|1Pondo|heydouga|pacopacomama|muramura|Tokyo.*Hot)[-_\s]`).
		ReplaceAllString(s, "") // trim prefixes
	s = regexp.MustCompile(`^(?i)\s*(FC2[-_]?PPV)[-_]`).
		ReplaceAllString(s, "FC2-") // normalize fc2 prefixes
	s = s[:findFirstNonASCII(s)] // trim unicode content
	s = strings.Fields(s)[0]     // trim possible alpha started title
	for re := regexp.MustCompile(`(?i)([-_](c|ch|cd\d{1,2})|ch)\s*$`); re.MatchString(s); {
		s = re.ReplaceAllString(s, "") // repeatedly trim suffixes
	}
	return strings.TrimSpace(s)
}

// IsUncensored returns true if the number is belonged to uncensored movie.
// It should be noted that this function is not accurate and can only be
// used to detect number of some certain movie studio or publisher.
func IsUncensored(s string) bool {
	return regexp.
		MustCompile(`^(?i)[\d-]{4,}|\d{6}_\d{2,3}|(cz|gedo|k|n|kb|red-|se)\d{2,4}|heyzo.+|fc2-.+|xxx-av-.+|heydouga-.+$`).
		MatchString(s)
}

// Similarity returns the similarity between two numbers.
func Similarity(a, b string) float64 {
	m := metrics.NewLevenshtein()
	m.CaseSensitive = false
	m.InsertCost = 1
	m.DeleteCost = 1
	m.ReplaceCost = 2
	return strutil.Similarity(a, b, m)
}

// RequireFaceDetection returns true if the movie cover
// requires face detection.
func RequireFaceDetection(s string) bool {
	if IsUncensored(s) {
		return true
	}
	if regexp.MustCompile(`(?i)^\d+[a-z]+`).MatchString(s) {
		return true
	}
	if regexp.MustCompile(`(?i)^(fcp|siro|msfh)`).MatchString(s) {
		return true
	}
	return false
}
