package site

import (
	goio "io"

	"jacobo.tarrio.org/jtweb/renderer/templates"
)

func (c *Contents) outputToc(w goio.Writer, lang string, names []string, tag string) error {
	t, err := templates.GetTemplates(c.Config, lang)
	if err != nil {
		return err
	}
	tmpl, err := t.GetTocTemplate(lang)
	if err != nil {
		return err
	}

	stories := make([]templates.PageData, len(names))
	for i, name := range names {
		stories[i] = c.makePageData(c.Pages[name])
	}

	tocData := templates.TocData{
		Tag:        tag,
		TotalCount: len(names),
		Stories:    stories,
	}

	return tmpl.Execute(w, tocData)
}
