package templates

import (
	"fmt"
	"html/template"
	"jtweb/languages"
	"jtweb/uri"
	"path/filepath"
	"time"
)

// Templates holds the configuration for the template system.
type Templates struct {
	TemplatePath      string
	WebRoot           string
	Site              LinkData
	tocTemplates      map[string]*template.Template
	pageTemplates     map[string]*template.Template
	indexTocTemplates map[string]*template.Template
}

// LinkData holds information about a link.
type LinkData struct {
	Name string
	URI  string
}

// PageData holds page information to be rendered.
type PageData struct {
	Title        string
	Permalink    string
	Author       LinkData
	Summary      string
	PublishDate  time.Time
	Tags         []string
	Content      template.HTML
	NewerPage    LinkData
	OlderPage    LinkData
	Translations []TranslationData
}

// TranslationData holds information about a translation.
type TranslationData struct {
	Name     string
	URI      string
	Language string
}

// TocData holds table-of-contents information to be rendered.
type TocData struct {
	BaseURI        string
	Tag            string
	Year           int
	YearCount      int
	TotalCount     int
	Stories        []PageData
	StoryYears     []int
	UndatedStories bool
}

// IndexTocData holds per-year TOC index data to be rendered.
type IndexTocData struct {
	BaseURI    string
	Tag        string
	TotalCount int
	Years      []YearData
}

// YearData holds one year's TOC information.
type YearData struct {
	Year  int
	Count int
	Tags  []string
}

// GetTocTemplate loads the table-of-contents template for a particular language.
func (t Templates) GetTocTemplate(lang string) (tmpl *template.Template, err error) {
	t.tocTemplates, tmpl, err = t.getTemplate(t.tocTemplates, "toc", lang)
	return
}

// GetPageTemplate loads the page template for a particular language.
func (t Templates) GetPageTemplate(lang string) (tmpl *template.Template, err error) {
	t.pageTemplates, tmpl, err = t.getTemplate(t.pageTemplates, "page", lang)
	return
}

// GetIndexTocTemplate loads the story index template for a particular language.
func (t Templates) GetIndexTocTemplate(lang string) (tmpl *template.Template, err error) {
	t.indexTocTemplates, tmpl, err = t.getTemplate(t.indexTocTemplates, "index-toc", lang)
	return
}

func (t Templates) getTemplate(cache map[string]*template.Template, name string, lang string) (map[string]*template.Template, *template.Template, error) {
	var err error
	if cache == nil {
		cache = make(map[string]*template.Template)
	}
	tmpl, ok := cache[lang]
	if !ok {
		tmpl, err = t.loadTemplate(name+"-"+lang+".tmpl", lang)
		if err != nil {
			return cache, nil, err
		}
		cache[lang] = tmpl
	}
	return cache, tmpl, nil
}

func (t Templates) loadTemplate(fileName string, lang string) (*template.Template, error) {
	dateLocale, err := languages.FindByCode(lang)
	if err != nil {
		return nil, err
	}
	return template.New(fileName).Funcs(template.FuncMap{
		"formatDate": func(tm time.Time) string {
			return t.FormatDate(tm, dateLocale)
		},
		"getTagTocURI": func(tag string) string {
			return t.GetURI(fmt.Sprintf("/tags/toc-%s-%s.html", uri.GetTagPath(tag), lang))
		},
		"getTagURIWithTime": func(tag string, tm time.Time) string {
			year := tm.Year()
			if tm.IsZero() {
				year = 0
			}
			return t.GetURI(fmt.Sprintf("/tags/%s-%s-%d.html", uri.GetTagPath(tag), lang, year))
		},
		"getTagURIWithYear": func(tag string, year int) string {
			return t.GetURI(fmt.Sprintf("/tags/%s-%s-%d.html", uri.GetTagPath(tag), lang, year))
		},
		"getTocURI": func() string {
			return t.GetURI(fmt.Sprintf("/toc/toc-%s.html", lang))
		},
		"getTocURIWithTime": func(tm time.Time) string {
			year := tm.Year()
			if tm.IsZero() {
				year = 0
			}
			return t.GetURI(fmt.Sprintf("/toc/toc-%s-%d.html", lang, year))
		},
		"getTocURIWithYear": func(year int) string {
			return t.GetURI(fmt.Sprintf("/toc/toc-%s-%d.html", lang, year))
		},
		"getURI": t.GetURI,
		"language": func() string {
			return lang
		},
		"plural": t.Plural,
		"site": func() LinkData {
			return t.Site
		},
		"webRoot": func() string {
			return t.WebRoot
		},
		"year": t.Year,
	}).ParseFiles(filepath.Join(t.TemplatePath, fileName))
}

// FormatDate renders the given time as a date.
func (t Templates) FormatDate(tm time.Time, locale languages.Language) string {
	if tm.IsZero() {
		return ""
	}
	return locale.FormatDate(tm)
}

// GetURI returns a path that's relative to the web root.
func (t Templates) GetURI(path string) string {
	return uri.Concat(t.WebRoot, path)
}

// Plural returns the singular or plural form depending on the value of the count.
func (t Templates) Plural(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

// Year returns the year part of the given time.
func (t Templates) Year(tm time.Time) int {
	return tm.Year()
}
