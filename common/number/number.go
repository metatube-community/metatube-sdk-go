package number

import (
	"path"
	"regexp"
	"strings"
)

func Trim(s string) string {
	const maxExtLength = 7
	if ext := path.Ext(s); len(ext) < maxExtLength {
		s = s[:len(s)-len(ext)] // trim extension
	}
	s = regexp.MustCompile(`(?i)([a-z\d]+\.(?:com|net|top|xyz|tv))(?:[^a-z\d]|$)`).
		ReplaceAllString(s, "") // trim domain
	if ss := regexp.MustCompile(`(?i)([a-z\d]+(?:[-_][a-z\d]{2,})+)`).FindStringSubmatch(s); len(ss) > 0 {
		s = ss[1] // first find number with dashes
	} else if ss = regexp.MustCompile(`(?i)((?:[a-z]+\d|\d+[a-z])[a-z\d]+)`).FindStringSubmatch(s); len(ss) > 1 {
		s = ss[1] // otherwise find number with alphas & digits
	}
	s = regexp.MustCompile(`(?i)^(?:f?hd|sd)[-_](.*$)`).
		ReplaceAllString(s, "${1}") // trim special prefixes
	s = regexp.MustCompile(`(?i)[-_.](dvd|iso|mkv|mp4|c?avi|\d*fps|whole|(f|hhb)?hd\d*|sd\d*|(?:360|480|720|1080|2160)[pi]|X1080X|uncensored|leak|[2468]k|[xh]26[45])+`).
		ReplaceAllString(s, "") // trim tags
	s = regexp.MustCompile(`(?i)(^|[-_\s]+)(carib(b?ean)?(com)?(pr)?|1?Pond?o?|10mu(sume)?|paco(paco)?(mama)?|mura(mura)?|Tokyo[-_\s]?Hot)([-_\s]+(?P<pattern>\d{4,}[-_]\d{2,}|[a-z]{1,4}\d{2,4})|$)`).
		ReplaceAllString(s, "${pattern}") // trim makers
	s = regexp.MustCompile(`^(?i)\s*(FC2[-_]?PPV)[-_]`).
		ReplaceAllString(s, "FC2-") // normalize fc2 prefixes
	for re := regexp.MustCompile(`(?i)([-_](c|uc|ch|cd\d{1,2})|ch|A|B|C|D)\s*$`); re.MatchString(s); {
		s = re.ReplaceAllString(s, "") // repeatedly trim suffixes
	}
	return strings.TrimSpace(s)
}

// IsUncensored returns true if the number is belonged to uncensored movie.
// It should be noted that this function is not accurate and can only be
// used to detect number of some certain movie studio.
func IsUncensored(s string) bool {
	return regexp.
		MustCompile(`^(?i)(\d{4,6}[-_]\d{2,3}|(cz|gedo|k|n|kb|se)\d{2,4}|(heyzo|xxx-av|heydouga|kin8)[-_].+)|([hc]0930|h4610|av9898|1000giri)[-_][a-z\d]+$`).
		MatchString(s)
}

// IsFC2 returns true if the number is fc2 video type.
func IsFC2(s string) bool {
	return regexp.
		MustCompile(`^(?i)FC2([-_]?PPV)?[-_]?\d+$`).
		MatchString(s)
}

// IsSpecial returns true if the number is special compare to other regular numbers.
func IsSpecial(s string) bool {
	if IsUncensored(s) || IsFC2(s) {
		return true
	}
	return regexp.
		MustCompile(`^(?i)(gcolle|getchu|gyutto|pcolle|mywife)[-_]?.+$`).
		MatchString(s)
}

// RequiresFaceDetection returns true if the movie cover
// requires face detection.
func RequiresFaceDetection(s string) bool {
	if IsSpecial(s) {
		return true
	}
	if regexp.MustCompile(`(?i)^\d+[a-z]+`).MatchString(s) {
		return true
	}
	if shiroutoRe.MatchString(s) {
		return true
	}
	return false
}
