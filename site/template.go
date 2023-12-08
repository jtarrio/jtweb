package site

import (
	"html/template"

	"jacobo.tarrio.org/jtweb/io"
	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/page"
)

func (c *Contents) renderTemplate(name string) error {
	source := c.Config.Files().Content().GoTo(name + ".tmpl")
	content, err := source.ReadBytes()
	if err != nil {
		return err
	}
	templateName := source.BaseName()
	tmpl, err := template.New(templateName).Funcs(template.FuncMap{
		"hasContent": func(lang string) bool {
			toc, ok := c.Toc[languages.FindByCodeWithFallback(lang, languages.LanguageEn)]
			if !ok {
				return false
			}
			return len(toc.All) > 0
		},
		"latestPage": func(lang string) *page.Page {
			toc, ok := c.Toc[languages.FindByCodeWithFallback(lang, languages.LanguageEn)]
			if !ok {
				return nil
			}
			name := toc.All[0]
			return c.Pages[name]
		},
		"webRoot": func(lang string) string {
			return c.Config.Site(languages.FindByCodeWithFallback(lang, languages.LanguageEn)).WebRoot()
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
