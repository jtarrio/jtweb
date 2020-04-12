package languages

import (
	"fmt"
	"time"
)

type languageEn struct{ languageBase }

// LanguageEn contains the English language definition.
var LanguageEn = &languageEn{languageBase{"English", "en", []string{"es", "gl"}}}

var longMonthsEn []string = []string{
	"January", "February", "March", "April",
	"May", "June", "July", "August",
	"September", "October", "November", "December",
}

func (l *languageEn) FormatDate(t time.Time) string {
	return fmt.Sprintf("%s %d, %d", longMonthsEn[t.Month()-1], t.Day(), t.Year())
}
