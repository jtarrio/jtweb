package site

import (
	"html/template"
	goio "io"
	"strings"

	"jacobo.tarrio.org/jtweb/page"
	"jacobo.tarrio.org/jtweb/renderer/templates"
	"jacobo.tarrio.org/jtweb/site/io"
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

func (c *Contents) OutputAsPage(w goio.Writer, t *templates.Templates, page *page.Page) error {
	tmpl, err := t.GetPageTemplate(page.Header.Language)
	if err != nil {
		return err
	}
	return c.runTemplate(tmpl, w, page)
}

func (c *Contents) OutputAsEmail(w goio.Writer, t *templates.Templates, page *page.Page) error {
	tmpl, err := t.GetEmailTemplate(page.Header.Language)
	if err != nil {
		return err
	}
	return c.runTemplate(tmpl, w, page)
}

func (c *Contents) OutputAsPlainEmail(w goio.Writer, t *templates.Templates, page *page.Page) error {
	tmpl, err := t.GetPlainEmailTemplate(page.Header.Language)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, c.makePageData(page))
}

func (c *Contents) runTemplate(tmpl *template.Template, w goio.Writer, page *page.Page) error {
	sb := strings.Builder{}
	err := tmpl.Execute(&sb, c.makePageData(page))
	if err != nil {
		return err
	}
	return templates.MakeUrisAbsolute(strings.NewReader(sb.String()), w, c.Config.GetWebRoot(page.Header.Language), page.Name)
}

func (c *Contents) makePageData(page *page.Page) templates.PageData {
	sb := strings.Builder{}
	page.Render(&sb)

	pageData := templates.PageData{
		Title:     page.Header.Title,
		Permalink: c.makePageURI(page),
		Author:    templates.LinkData{},
		Summary:   page.Header.Summary,
		Episode:   page.Header.Episode,
		Tags:      page.Header.Tags,
		Content:   template.HTML(sb.String()),
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
			pageData.Author.Name = c.Config.GetAuthorName()
			pageData.Author.URI = c.Config.GetAuthorURI()
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
			templates.TranslationData{
				Name:     translation.Header.Title,
				URI:      c.makePageURI(translation),
				Language: t.Language,
			})
	}
	return pageData
}

func (c *Contents) makePageURI(p *page.Page) string {
	return uri.Concat(c.Config.GetWebRoot(p.Header.Language), p.Name+".html")
}
