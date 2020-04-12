package languages

import (
	"fmt"
	"time"
)

type languageGl struct{ languageBase }

// LanguageGl contains the Galician language definition.
var LanguageGl = &languageGl{languageBase{"galego", "gl", []string{"es", "en"}}}

var longMonthsGl []string = []string{
	"xaneiro", "febreiro", "marzo", "abril",
	"maio", "xuño", "xullo", "agosto",
	"setembro", "outubro", "novembro", "decembro",
}

func (l *languageGl) FormatDate(t time.Time) string {
	return fmt.Sprintf("%d de %s de %d", t.Day(), longMonthsGl[t.Month()-1], t.Year())
}
