package uri

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// GetTagPath returns a path component for a tag, where only latin and numeric characters are kept.
func GetTagPath(str string) string {
	sep := false
	var sb strings.Builder
	for _, c := range removeMarks(str) {
		c = unicode.ToLower(c)
		if (c >= rune('0') && c <= rune('9')) || (c >= rune('a') && c <= rune('z')) {
			if sep {
				sb.WriteByte('_')
				sep = false
			}
			sb.WriteRune(c)
		} else {
			sep = true
		}
	}
	return sb.String()
}

// Concat adds a path component to a URI.
func Concat(components ...string) string {
	var sb strings.Builder
	prevSlash := false
	for _, c := range components {
		if c == "" {
			continue
		}
		cSlash := strings.HasPrefix(c, "/")
		if sb.Len() == 0 {
			sb.WriteString(c)
		} else if prevSlash != cSlash {
			sb.WriteString(c)
		} else if prevSlash {
			sb.WriteString(c[1:])
		} else {
			sb.WriteRune('/')
			sb.WriteString(c)
		}
		prevSlash = strings.HasSuffix(c, "/")
	}
	return sb.String()
}

func removeMarks(str string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	o := make([]byte, len(str)*2)
	n, _, err := t.Transform(o, []byte(str), true)
	if err != nil {
		return str
	}
	return string(o[:n])
}
