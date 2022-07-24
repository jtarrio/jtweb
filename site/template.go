package site

import (
	"html/template"
	"io"
	"path/filepath"

	"jacobo.tarrio.org/jtweb/page"
)

func (c *Contents) renderTemplate(name string) error {
	inputFileName := filepath.Join(c.InputPath, name+".tmpl")
	templateName := filepath.Base(inputFileName)
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
		"webRoot": func() string {
			return c.WebRoot
		},
	}).ParseFiles(inputFileName)
	if err != nil {
		return err
	}
	return makeFile(
		filepath.Join(c.OutputPath, name),
		func(w io.Writer) error {
			return tmpl.Execute(w, c)
		})
}
