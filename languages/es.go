package languages

import (
	"fmt"
	"time"
)

type languageEs struct{ languageBase }

// LanguageEs contains the Spanish language definition.
var LanguageEs = &languageEs{languageBase{"español", "es", []string{"gl", "en"}}}

var longMonthsEs []string = []string{
	"enero", "febrero", "marzo", "abril",
	"mayo", "junio", "julio", "agosto",
	"setiembre", "octubre", "noviembre", "diciembre",
}

func (l *languageEs) FormatDate(t time.Time) string {
	return fmt.Sprintf("%d de %s de %d", t.Day(), longMonthsEs[t.Month()-1], t.Year())
}
