package templates

import (
	"fmt"
	"html/template"
	"strings"
	textTemplate "text/template"
	"time"

	"jacobo.tarrio.org/jtweb/config"
	"jacobo.tarrio.org/jtweb/io"
	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/uri"
)

// Templates holds the configuration for the template system.
type Templates struct {
	err           error
	config        config.SiteConfig
	templateBase  io.File
	templates     map[string]*template.Template
	textTemplates map[string]*textTemplate.Template
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
	Translations []*TranslationData
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
	Stories    []*PageData
}

// GetTemplates returns a loader for templates for a particular language.
func GetTemplates(c config.Config, lang languages.Language) *Templates {
	return &Templates{
		config:        c.Site(lang),
		templateBase:  c.Files().Templates(),
		templates:     make(map[string]*template.Template),
		textTemplates: make(map[string]*textTemplate.Template),
	}
}

// Toc loads the table-of-contents template.
func (t *Templates) Toc() (*template.Template, error) {
	return t.getTemplate("toc")
}

// Page loads the page template.
func (t *Templates) Page() (*template.Template, error) {
	return t.getTemplate("page")
}

// GetPageTemplate loads the email template.
func (t *Templates) Email() (*template.Template, error) {
	return t.getTemplate("email")
}

// GetPageTemplate loads the plain-txt email template.
func (t Templates) PlainEmail() (*textTemplate.Template, error) {
	return t.getTextTemplate("email-plain")
}

// IndexToc loads the story index template.
func (t Templates) IndexToc() (*template.Template, error) {
	return t.getTemplate("index-toc")
}

func (t *Templates) getTemplate(name string) (*template.Template, error) {
	if t.err != nil {
		return nil, t.err
	}
	out, ok := t.templates[name]
	if ok {
		return out, nil
	}

	fileName := name + "-" + t.getLanguage() + ".tmpl"
	tmpl, err := t.templateBase.GoTo(fileName).ReadBytes()
	if err != nil {
		return nil, err
	}
	out, err = template.New(fileName).Funcs(template.FuncMap{
		"formatDate": t.formatDate,
		"getTagURI":  t.getTagURI,
		"getTocURI":  t.getTocURI,
		"getURI":     t.getURI,
		"language":   t.getLanguage,
		"plural":     t.plural,
		"site":       t.getSite,
		"webRoot":    t.getWebroot,
	}).Parse(string(tmpl))
	if err != nil {
		return nil, err
	}
	t.templates[name] = out
	return out, nil
}

func (t *Templates) getTextTemplate(name string) (*textTemplate.Template, error) {
	if t.err != nil {
		return nil, t.err
	}
	out, ok := t.textTemplates[name]
	if ok {
		return out, nil
	}

	fileName := name + "-" + t.getLanguage() + ".tmpl"
	tmpl, err := t.templateBase.GoTo(fileName).ReadBytes()
	if err != nil {
		return nil, err
	}
	out, err = textTemplate.New(fileName).Funcs(textTemplate.FuncMap{
		"formatDate": t.formatDate,
		"getTagURI":  t.getTagURI,
		"getTocURI":  t.getTocURI,
		"getURI":     t.getURI,
		"htmlToText": t.htmlToText,
		"language":   t.getLanguage,
		"plural":     t.plural,
		"site":       t.getSite,
		"webRoot":    t.getWebroot,
	}).Parse(string(tmpl))
	if err != nil {
		return nil, err
	}
	t.textTemplates[name] = out
	return out, nil
}

// FormatDate renders the given time as a date.
func (t *Templates) formatDate(tm time.Time) string {
	if tm.IsZero() {
		return ""
	}
	return t.config.Language().FormatDate(tm)
}

func (t *Templates) getTagURI(tag string) string {
	return t.getURI(fmt.Sprintf("/tags/%s-%s.html", uri.GetTagPath(tag), t.getLanguage()))
}

func (t *Templates) getTocURI() string {
	return t.getURI(fmt.Sprintf("/toc/toc-%s.html", t.getLanguage()))
}

// getURI returns a path that's relative to the web root.
func (t *Templates) getURI(path string) string {
	return uri.Concat(t.getWebroot(), path)
}

func (t *Templates) getLanguage() string {
	return t.config.Language().Code()
}

// plural returns the singular or plural form depending on the value of the count.
func (t *Templates) plural(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func (t *Templates) htmlToText(content template.HTML, linksTitle string, pictureTitle string) string {
	sb := strings.Builder{}
	err := HtmlToText(strings.NewReader(string(content)), &sb, linksTitle, pictureTitle)
	if err != nil {
		panic(err)
	}
	return sb.String()
}

func (t *Templates) getSite() LinkData {
	return LinkData{
		Name: t.config.Name(),
		URI:  t.config.Uri(),
	}
}

func (t *Templates) getWebroot() string {
	return t.config.WebRoot()
}
