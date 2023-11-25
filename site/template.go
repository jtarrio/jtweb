package site

import (
	"html/template"

	"jacobo.tarrio.org/jtweb/page"
	"jacobo.tarrio.org/jtweb/site/io"
)

func (c *Contents) renderTemplate(name string) error {
	source := c.Config.GetInputBase().GoTo(name + ".tmpl")
	content, err := source.ReadBytes()
	if err != nil {
		return err
	}
	templateName := source.BaseName()
	tmpl, err := template.New(templateName).Funcs(template.FuncMap{
		"hasContent": func(lang string) bool {
			toc, ok := c.Toc[lang]
			if !ok {
				return false
			}
			return len(toc.All) > 0
		},
		"latestPage": func(lang string) *page.Page {
			toc, ok := c.Toc[lang]
			if !ok {
				return nil
			}
			name := toc.All[0]
			return c.Pages[name]
		},
		"webRoot": func(lang string) string {
			return c.Config.GetWebRoot(lang)
		},
	}).Parse(string(content))
	if err != nil {
		return err
	}
	return c.makeFile(
		name,
		func(w io.Output) error {
			return tmpl.Execute(w, c)
		})
}
