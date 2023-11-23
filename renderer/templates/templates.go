package templates

import (
	"fmt"
	"html/template"
	"strings"
	text_template "text/template"
	"time"

	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/site/io"
	"jacobo.tarrio.org/jtweb/uri"
)

// Templates holds the configuration for the template system.
type Templates struct {
	TemplateBase        io.File
	WebRoot             string
	Site                LinkData
	tocTemplates        map[string]*template.Template
	pageTemplates       map[string]*template.Template
	emailTemplates      map[string]*template.Template
	plainEmailTemplates map[string]*text_template.Template
	indexTocTemplates   map[string]*template.Template
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
	Episode      string
	PublishDate  time.Time
	Tags         []string
	Content      template.HTML
	NewerPage    LinkData
	OlderPage    LinkData
	Translations []TranslationData
	Draft        bool
}

// TranslationData holds information about a translation.
type TranslationData struct {
	Name     string
	URI      string
	Language string
}

// TocData holds table-of-contents information to be rendered.
type TocData struct {
	Tag        string
	TotalCount int
	Stories    []PageData
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

// GetPageTemplate loads the email template for a particular language.
func (t Templates) GetEmailTemplate(lang string) (tmpl *template.Template, err error) {
	t.emailTemplates, tmpl, err = t.getTemplate(t.emailTemplates, "email", lang)
	return
}

// GetPageTemplate loads the plain-txt email template for a particular language.
func (t Templates) GetPlainEmailTemplate(lang string) (tmpl *text_template.Template, err error) {
	t.plainEmailTemplates, tmpl, err = t.getTextTemplate(t.plainEmailTemplates, "email-plain", lang)
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

func (t Templates) getTextTemplate(cache map[string]*text_template.Template, name string, lang string) (map[string]*text_template.Template, *text_template.Template, error) {
	var err error
	if cache == nil {
		cache = make(map[string]*text_template.Template)
	}
	tmpl, ok := cache[lang]
	if !ok {
		tmpl, err = t.loadTextTemplate(name+"-"+lang+".tmpl", lang)
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
	tmpl, err := t.TemplateBase.GoTo(fileName).ReadBytes()
	if err != nil {
		return nil, err
	}
	return template.New(fileName).Funcs(template.FuncMap{
		"formatDate": func(tm time.Time) string {
			return t.FormatDate(tm, dateLocale)
		},
		"getTagURI": func(tag string) string {
			return t.GetURI(fmt.Sprintf("/tags/%s-%s.html", uri.GetTagPath(tag), lang))
		},
		"getTocURI": func() string {
			return t.GetURI(fmt.Sprintf("/toc/toc-%s.html", lang))
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
	}).Parse(string(tmpl))
}

func (t Templates) loadTextTemplate(fileName string, lang string) (*text_template.Template, error) {
	dateLocale, err := languages.FindByCode(lang)
	if err != nil {
		return nil, err
	}
	tmpl, err := t.TemplateBase.GoTo(fileName).ReadBytes()
	if err != nil {
		return nil, err
	}
	return text_template.New(fileName).Funcs(text_template.FuncMap{
		"formatDate": func(tm time.Time) string {
			return t.FormatDate(tm, dateLocale)
		},
		"getTagURI": func(tag string) string {
			return t.GetURI(fmt.Sprintf("/tags/%s-%s.html", uri.GetTagPath(tag), lang))
		},
		"getTocURI": func() string {
			return t.GetURI(fmt.Sprintf("/toc/toc-%s.html", lang))
		},
		"getURI":     t.GetURI,
		"htmlToText": t.HtmlToText,
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
	}).ParseFiles(string(tmpl))
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

func (t Templates) HtmlToText(content template.HTML, linksTitle string, pictureTitle string) string {
	sb := strings.Builder{}
	err := HtmlToText(strings.NewReader(string(content)), &sb, linksTitle, pictureTitle)
	if err != nil {
		panic(err)
	}
	return sb.String()
}
