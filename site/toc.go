package site

import (
	goio "io"

	"jacobo.tarrio.org/jtweb/renderer/templates"
)

func (c *Contents) outputToc(w goio.Writer, lang string, names []string, tag string) error {
	tmpl, err := templates.GetTemplates(c.Config, lang).Toc()
	if err != nil {
		return err
	}

	stories := make([]*templates.PageData, len(names))
	for i, name := range names {
		pageData, err := c.makePageData(c.Pages[name])
		if err != nil {
			return err
		}
		stories[i] = pageData
	}

	tocData := templates.TocData{
		Tag:        tag,
		TotalCount: len(names),
		Stories:    stories,
	}

	return tmpl.Execute(w, tocData)
}
