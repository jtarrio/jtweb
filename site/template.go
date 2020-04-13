package site

import (
	"html/template"
	"io"
	"jtweb/page"
	"math"
	"path/filepath"
)

func (c *Contents) renderTemplate(name string) error {
	inputFileName := filepath.Join(c.InputPath, name+".tmpl")
	templateName := filepath.Base(inputFileName)
	tmpl, err := template.New(templateName).Funcs(template.FuncMap{
		"latestPage": func(lang string) *page.Page {
			toc, ok := c.Toc[lang]
			if !ok {
				return nil
			}
			years := toc.All.ByYear
			if len(years) == 0 {
				return nil
			}
			year := math.MinInt32
			for y := range years {
				if y > year {
					year = y
				}
			}
			name := years[year][0]
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
