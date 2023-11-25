package site

import (
	"fmt"
	goio "io"
	"path/filepath"
	"strings"

	"jacobo.tarrio.org/jtweb/page"
	"jacobo.tarrio.org/jtweb/renderer/templates"
	"jacobo.tarrio.org/jtweb/site/config"
	"jacobo.tarrio.org/jtweb/site/io"
	"jacobo.tarrio.org/jtweb/uri"
)

// Contents contains the parsed and indexed content of the site.
type Contents struct {
	Config       config.Config
	Files        []string
	Templates    []string
	Pages        map[string]*page.Page
	Tags         map[string]string
	Toc          GlobalTableOfContents
	Translations map[string][]Translation
}

// GlobalTableOfContents contains the tables of contents for every language.
type GlobalTableOfContents map[string]LanguageTableOfContents

// LanguageTableOfContents contains the tables of contents for every year in a language.
type LanguageTableOfContents struct {
	// TOC for all pages.
	All TableOfContents
	// TOC for each tag.
	ByTag map[string]TableOfContents
	// Map from each page name to the immediately newer page's name.
	NewerPages map[string]string
	// Map from each page name to the immediately older page's name.
	OlderPages map[string]string
}

// TableOfContents contains a list of pages.
type TableOfContents []string

// Translation contains information about a page translation.
type Translation struct {
	Name     string
	Language string
}

// Read parses the whole site contents.
func Read(s config.Config) (*Contents, error) {
	files := make([]string, 0)
	templates := make([]string, 0)
	pagesByName := make(map[string]*page.Page)
	tagNames := make(map[string]string)
	err := s.GetInputBase().ForAllFiles(func(file io.File, err error) error {
		if err != nil {
			return err
		}
		name := file.Name()
		if strings.HasSuffix(name, ".md") {
			page, err := parsePage(file)
			if err != nil {
				return fmt.Errorf("error parsing page %s: %v", file.Name(), err)
			}
			if page.Header.PublishDate.After(s.GetPublishUntil()) {
				return nil
			}
			pagesByName[page.Name] = page
			for _, tag := range page.Header.Tags {
				tagNames[uri.GetTagPath(tag)] = tag
			}
		} else if strings.HasSuffix(name, ".tmpl") {
			templates = append(templates, name[:len(name)-5])
		} else {
			files = append(files, name)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	translationsByName, err := getTranslationsByName(pagesByName)
	if err != nil {
		return nil, err
	}

	tocByLanguage, err := indexPages(pagesByName, translationsByName, s.GetHideUntranslated())
	if err != nil {
		return nil, err
	}

	c := Contents{
		Config:       s,
		Files:        files,
		Templates:    templates,
		Pages:        pagesByName,
		Tags:         tagNames,
		Toc:          tocByLanguage,
		Translations: translationsByName,
	}
	return &c, nil
}

type filePopulator func(w io.Output) error

// Write converts the site contents to HTML and writes it to disk.
func (c *Contents) Write() error {
	for _, file := range c.Files {
		if file == ".htaccess" {
			err := c.outputHtaccess(file)
			if err != nil {
				return err
			}
		} else if filepath.Base(file)[0] == '.' {
			continue
		} else {
			err := c.copyFile(file)
			if err != nil {
				return err
			}
		}
	}
	for _, name := range c.Templates {
		err := c.renderTemplate(name)
		if err != nil {
			return err
		}
	}
	for _, page := range c.Pages {
		t := &templates.Templates{
			TemplateBase: c.Config.GetTemplateBase(),
			WebRoot:      c.Config.GetWebRoot(page.Header.Language),
			Site: templates.LinkData{
				Name: c.Config.GetSiteName(page.Header.Language),
				URI:  c.Config.GetSiteURI(page.Header.Language),
			},
		}
		err := c.makeFile(
			page.Name+".html",
			func(w io.Output) error {
				return c.outputPage(w, t, page)
			})
		if err != nil {
			return fmt.Errorf("error rendering page %s: %v", page.Name, err)
		}
	}
	for lang, languageToc := range c.Toc {
		t := &templates.Templates{
			TemplateBase: c.Config.GetTemplateBase(),
			WebRoot:      c.Config.GetWebRoot(lang),
			Site: templates.LinkData{
				Name: c.Config.GetSiteName(lang),
				URI:  c.Config.GetSiteURI(lang),
			},
		}
		err := c.makeFile(
			fmt.Sprintf("toc/toc-%s.html", lang),
			func(w io.Output) error {
				return c.outputToc(w, t, lang, languageToc.All, "")
			})
		if err != nil {
			return err
		}
		for tag, tagToc := range languageToc.ByTag {
			err := c.makeFile(
				fmt.Sprintf("tags/%s-%s.html", tag, lang),
				func(w io.Output) error {
					return c.outputToc(w, t, lang, tagToc, c.Tags[tag])
				})
			if err != nil {
				return err
			}
		}
		err = c.makeFile(
			fmt.Sprintf("rss/%s.xml", lang),
			func(w io.Output) error {
				return c.outputRss(w, t, lang)
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Contents) copyFile(name string) error {
	in := c.Config.GetInputBase().GoTo(name)
	out := c.Config.GetOutputBase().GoTo(name)
	stat, err := in.Stat()
	if err != nil {
		return err
	}
	input, err := in.Read()
	if err != nil {
		return err
	}
	output, err := out.Create()
	if err != nil {
		input.Close()
		return err
	}
	_, err = goio.Copy(output, input)
	input.Close()
	output.Close()
	if err != nil {
		return err
	}
	return out.Chtime(stat.ModTime)
}

func (c *Contents) makeFile(path string, populator filePopulator) error {
	file := c.Config.GetOutputBase().GoTo(path)
	output, err := file.Create()
	if err != nil {
		return err
	}
	defer output.Close()
	err = populator(output)
	return err
}
