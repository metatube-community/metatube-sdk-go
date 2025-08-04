package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"golang.org/x/net/html"
	dt "gorm.io/datatypes"

	"github.com/metatube-community/metatube-sdk-go/common/convertor"
)

// ParseInt parses string to int regardless.
func ParseInt(s string) int {
	s = strings.TrimSpace(s)
	n, _ := strconv.Atoi(s)
	return n
}

// ParseTime parses a string with a valid time format into time.Time.
func ParseTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if ss := regexp.MustCompile(`([\s\d]+)年([\s\d]+)月([\s\d]+)日`).
		FindStringSubmatch(s); len(ss) == 4 {
		s = fmt.Sprintf("%s-%s-%s",
			strings.TrimSpace(ss[1]),
			strings.TrimSpace(ss[2]),
			strings.TrimSpace(ss[3]))
	}
	t, _ := dateparse.ParseAny(s)
	return t
}

// ParseDate parses a string with a valid date format into Date.
func ParseDate(s string) dt.Date {
	return dt.Date(ParseTime(s))
}

// ParseDuration parses a string with valid duration format into time.Duration.
func ParseDuration(s string) time.Duration {
	s = convertor.ReplaceSpaceAll(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "秒", "s")
	s = strings.ReplaceAll(s, "分", "m")
	s = strings.ReplaceAll(s, "時", "h")
	s = strings.ReplaceAll(s, "时", "h")
	s = strings.ReplaceAll(s, "sec", "s")
	s = strings.ReplaceAll(s, "min", "m")
	if ss := regexp.MustCompile(`(?i)(\d+):(\d+):(\d+)`).FindStringSubmatch(s); len(ss) > 0 {
		s = fmt.Sprintf("%02sh%02sm%02ss", ss[1], ss[2], ss[3])
	} else if ss := regexp.MustCompile(`(?i)(\d+):(\d+)`).FindStringSubmatch(s); len(ss) > 0 {
		s = fmt.Sprintf("%02sm%02ss", ss[1], ss[2])
	} else if ss := regexp.MustCompile(`(?i)(\d+[mhs]?)`).FindAllStringSubmatch(s, -1); len(ss) > 0 {
		ds := make([]string, 0, 3)
		for _, d := range ss {
			ds = append(ds, d[1])
		}
		s = strings.Join(ds, "")
	}
	d, _ := time.ParseDuration(s)
	return d
}

// ParseRuntime parses a string into time.Duration and converts it to minutes as integer.
func ParseRuntime(s string) int {
	return int(ParseDuration(s).Minutes())
}

// ParseScore parses a string into a float-based score.
func ParseScore(s string) float64 {
	s = strings.ReplaceAll(s, "点", "")
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return 0
	}
	s = strings.TrimSpace(fields[0])
	n, _ := strconv.ParseFloat(s, 64)
	return n
}

// ParseTexts parses all plaintext from the given *html.Node.
func ParseTexts(n *html.Node, texts *[]string) {
	if n.Type == html.TextNode {
		if text := strings.TrimSpace(n.Data); text != "" {
			*texts = append(*texts, text)
		}
	}
	for n := n.FirstChild; n != nil; n = n.NextSibling {
		ParseTexts(n, texts)
	}
}

func ParseActorNames(s string) (names []string) {
	add := func(name string) {
		if name = strings.TrimSpace(name); len(name) > 0 {
			names = append(names, name)
		}
	}
	sb := &strings.Builder{}
	for _, r := range s {
		switch r {
		case '、', ';', ',':
			fallthrough
		case '(', '（':
			fallthrough
		case ')', '）':
			add(sb.String())
			sb.Reset()
		default:
			sb.WriteRune(r)
		}
	}
	add(sb.String())
	return
}

func ParseIDToNumber(s string) string {
	s = strings.ToUpper(s)
	if ss := regexp.MustCompile(`(\d*[A-Z]+)(\d+)`).FindStringSubmatch(s); len(ss) >= 3 {
		return fmt.Sprintf("%s-%s", ss[1], ss[2])
	}
	return s
}

func ParseBustCupSize(s string) (int, string, error) {
	sizeRe := regexp.MustCompile(`^(\d+)\s?([A-Z])$`)
	match := sizeRe.FindStringSubmatch(s)

	if len(match) != 3 {
		return 0, "", fmt.Errorf("invalid format: %s", s)
	}

	num := match[1]
	unit := match[2]

	value, err := strconv.Atoi(num)
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse numeric part '%s': %w", num, err)
	}
	return value, unit, nil
}
