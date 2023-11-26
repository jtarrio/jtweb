package templates

import (
	"fmt"
	"html/template"
	"strings"
	text_template "text/template"
	"time"

	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/site/config"
	"jacobo.tarrio.org/jtweb/site/io"
	"jacobo.tarrio.org/jtweb/uri"
)

// Templates holds the configuration for the template system.
type Templates struct {
	config             config.Config
	locale             languages.Language
	templateBase       io.File
	tocTemplate        *template.Template
	pageTemplate       *template.Template
	emailTemplate      *template.Template
	plainEmailTemplate *text_template.Template
	indexTocTemplate   *template.Template
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

func GetTemplates(c config.Config, lang string) (*Templates, error) {
	locale, err := languages.FindByCode(lang)
	if err != nil {
		return nil, err
	}
	return &Templates{
		config:             c,
		locale:             locale,
		templateBase:       c.GetTemplateBase(),
		tocTemplate:        nil,
		pageTemplate:       nil,
		emailTemplate:      nil,
		plainEmailTemplate: nil,
		indexTocTemplate:   nil,
	}, nil
}

// GetTocTemplate loads the table-of-contents template for a particular language.
func (t *Templates) GetTocTemplate(lang string) (*template.Template, error) {
	if t.tocTemplate == nil {
		template, err := t.getTemplate("toc")
		if err != nil {
			return nil, err
		}
		t.tocTemplate = template
	}
	return t.tocTemplate, nil
}

// GetPageTemplate loads the page template for a particular language.
func (t *Templates) GetPageTemplate(lang string) (*template.Template, error) {
	if t.pageTemplate == nil {
		template, err := t.getTemplate("page")
		if err != nil {
			return nil, err
		}
		t.pageTemplate = template
	}
	return t.pageTemplate, nil
}

// GetPageTemplate loads the email template for a particular language.
func (t *Templates) GetEmailTemplate(lang string) (*template.Template, error) {
	if t.emailTemplate == nil {
		template, err := t.getTemplate("email")
		if err != nil {
			return nil, err
		}
		t.emailTemplate = template
	}
	return t.emailTemplate, nil
}

// GetPageTemplate loads the plain-txt email template for a particular language.
func (t Templates) GetPlainEmailTemplate(lang string) (tmpl *text_template.Template, err error) {
	if t.plainEmailTemplate == nil {
		template, err := t.getTextTemplate("email-plain")
		if err != nil {
			return nil, err
		}
		t.plainEmailTemplate = template
	}
	return t.plainEmailTemplate, nil
}

// GetIndexTocTemplate loads the story index template for a particular language.
func (t Templates) GetIndexTocTemplate(lang string) (tmpl *template.Template, err error) {
	if t.indexTocTemplate == nil {
		template, err := t.getTemplate("index-toc")
		if err != nil {
			return nil, err
		}
		t.indexTocTemplate = template
	}
	return t.indexTocTemplate, nil
}

func (t *Templates) getTemplate(name string) (*template.Template, error) {
	fileName := name + "-" + t.locale.Code() + ".tmpl"
	tmpl, err := t.templateBase.GoTo(fileName).ReadBytes()
	if err != nil {
		return nil, err
	}
	return template.New(fileName).Funcs(template.FuncMap{
		"formatDate": t.formatDate,
		"getTagURI":  t.getTagURI,
		"getTocURI":  t.getTocURI,
		"getURI":     t.getURI,
		"language":   t.getLanguage,
		"plural":     t.plural,
		"site":       t.getSite,
		"webRoot":    t.getWebroot,
	}).Parse(string(tmpl))
}

func (t *Templates) getTextTemplate(name string) (*text_template.Template, error) {
	fileName := name + "-" + t.locale.Code() + ".tmpl"
	tmpl, err := t.templateBase.GoTo(fileName).ReadBytes()
	if err != nil {
		return nil, err
	}
	return text_template.New(fileName).Funcs(text_template.FuncMap{
		"formatDate": t.formatDate,
		"getTagURI":  t.getTagURI,
		"getTocURI":  t.getTocURI,
		"getURI":     t.getURI,
		"htmlToText": t.htmlToText,
		"language":   t.getLanguage,
		"plural":     t.plural,
		"site":       t.getSite,
		"webRoot":    t.getWebroot,
	}).ParseFiles(string(tmpl))
}

// FormatDate renders the given time as a date.
func (t *Templates) formatDate(tm time.Time) string {
	if tm.IsZero() {
		return ""
	}
	return t.locale.FormatDate(tm)
}

func (t *Templates) getTagURI(tag string) string {
	return t.getURI(fmt.Sprintf("/tags/%s-%s.html", uri.GetTagPath(tag), t.locale.Code()))
}

func (t *Templates) getTocURI() string {
	return t.getURI(fmt.Sprintf("/toc/toc-%s.html", t.locale.Code()))
}

// getURI returns a path that's relative to the web root.
func (t *Templates) getURI(path string) string {
	return uri.Concat(t.getWebroot(), path)
}

func (t *Templates) getLanguage() string {
	return t.locale.Code()
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
		Name: t.config.GetSiteName(t.locale.Code()),
		URI:  t.config.GetSiteURI(t.locale.Code()),
	}
}

func (t *Templates) getWebroot() string {
	return t.config.GetWebRoot(t.locale.Code())
}
