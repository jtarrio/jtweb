package site

import (
	"bytes"
	"html/template"
	goio "io"
	"strings"
	textTemplate "text/template"

	"github.com/aymerick/douceur/inliner"
	"jacobo.tarrio.org/jtweb/io"
	"jacobo.tarrio.org/jtweb/page"
	"jacobo.tarrio.org/jtweb/renderer"
	"jacobo.tarrio.org/jtweb/renderer/templates"
	"jacobo.tarrio.org/jtweb/uri"
)

func parsePage(file io.File) (*page.Page, error) {
	name := file.Name()
	name = name[:len(name)-3]
	input, err := file.Read()
	if err != nil {
		return nil, err
	}
	defer input.Close()
	return page.Parse(name, input)
}

func (c *Contents) OutputAsPage(w goio.Writer, page *page.Page) error {
	tmpl, err := templates.GetTemplates(c.Config, page.Header.Language).Page()
	if err != nil {
		return err
	}
	return c.outputPageFromTemplate(w, tmpl, page)
}

func (c *Contents) OutputAsEmail(w goio.Writer, page *page.Page) error {
	tmpl, err := templates.GetTemplates(c.Config, page.Header.Language).Email()
	if err != nil {
		return err
	}
	sb := strings.Builder{}
	err = c.outputPageFromTemplate(&sb, tmpl, page)
	if err != nil {
		return err
	}
	out, err := inliner.Inline(sb.String())
	if err != nil {
		return err
	}
	_, err = strings.NewReader(out).WriteTo(w)
	return err
}

func (c *Contents) OutputAsPlainEmail(w goio.Writer, page *page.Page) error {
	tmpl, err := templates.GetTemplates(c.Config, page.Header.Language).PlainEmail()
	if err != nil {
		return err
	}
	return c.outputPageFromTextTemplate(w, tmpl, page)
}

func (c *Contents) outputPageFromTemplate(w goio.Writer, tmpl *template.Template, page *page.Page) error {
	pageData, err := c.makePageData(page)
	if err != nil {
		return err
	}
	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, pageData)
	if err != nil {
		return err
	}
	return renderer.NormalizeOutput(w, &buf)
}

func (c *Contents) outputPageFromTextTemplate(w goio.Writer, tmpl *textTemplate.Template, page *page.Page) error {
	pageData, err := c.makePageData(page)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, pageData)
}

func (c *Contents) makePageData(page *page.Page) (*templates.PageData, error) {
	buf := bytes.Buffer{}
	err := page.Render(&buf)
	if err != nil {
		return nil, err
	}
	content := bytes.Buffer{}
	err = renderer.SanitizePost(&content, &buf, c.Config.Site(page.Header.Language).WebRoot(), string(page.Name))
	if err != nil {
		return nil, err
	}

	pageData := &templates.PageData{
		Title:     page.Header.Title,
		Permalink: c.makePageURI(page),
		Author:    templates.LinkData{},
		Summary:   page.Header.Summary,
		Episode:   page.Header.Episode,
		Tags:      page.Header.Tags,
		Content:   template.HTML(content.String()),
		Draft:     page.Header.Draft,
	}
	if !page.Header.HidePublishDate {
		pageData.PublishDate = page.Header.PublishDate
	}
	if !page.Header.HideAuthor {
		pageData.Author = templates.LinkData{
			Name: page.Header.AuthorName,
			URI:  page.Header.AuthorURI,
		}
		if pageData.Author.Name == "" && pageData.Author.URI == "" {
			pageData.Author.Name = c.Config.Author().Name()
			pageData.Author.URI = c.Config.Author().Uri()
		}
	}
	newer := c.Toc[page.Header.Language].NewerPages[page.Name]
	if newer != "" {
		pageData.NewerPage = templates.LinkData{
			URI:  c.makePageURI(c.Pages[newer]),
			Name: c.Pages[newer].Header.Title,
		}
	}
	older := c.Toc[page.Header.Language].OlderPages[page.Name]
	if older != "" {
		pageData.OlderPage = templates.LinkData{
			URI:  c.makePageURI(c.Pages[older]),
			Name: c.Pages[older].Header.Title,
		}
	}
	translations := c.Translations[page.Name]
	for _, t := range translations {
		translation := c.Pages[t.Name]
		pageData.Translations = append(
			pageData.Translations,
			&templates.TranslationData{
				Name:     translation.Header.Title,
				URI:      c.makePageURI(translation),
				Language: t.Language.Code(),
			})
	}
	return pageData, nil
}

func (c *Contents) makePageURI(p *page.Page) string {
	return uri.Concat(c.Config.Site(p.Header.Language).WebRoot(), string(p.Name)+".html")
}
