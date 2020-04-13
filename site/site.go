package site

import (
	"fmt"
	"io"
	"jtweb/page"
	"jtweb/renderer/templates"
	"jtweb/uri"
	"os"
	"path/filepath"
	"strings"
)

// Config contains the parameters of the current site.
type Config struct {
	TemplatePath string `yaml:"template_path"`
	InputPath    string `yaml:"input_path"`
	OutputPath   string `yaml:"output_path"`
	WebRoot      string `yaml:"webroot"`
	AuthorName   string `yaml:"author_name"`
	AuthorURI    string `yaml:"author_uri"`
}

// Contents contains the parsed and indexed content of the site.
type Contents struct {
	Config
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

// TableOfContents contains a list of pages for each year.
type TableOfContents struct {
	TotalCount int
	ByYear     map[int]SingleTableOfContents
}

// SingleTableOfContents contains a list of pages for a single year.
type SingleTableOfContents []string

// Translation contains information about a page translation.
type Translation struct {
	Name     string
	Language string
}

// Read parses the whole site contents.
func (s Config) Read() (*Contents, error) {
	files := make([]string, 0)
	templates := make([]string, 0)
	pagesByName := make(map[string]*page.Page)
	tagNames := make(map[string]string)
	err := filepath.Walk(s.InputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name, err := filepath.Rel(s.InputPath, path)
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".md") {
			page, err := parsePage(path, name)
			if err != nil {
				return fmt.Errorf("Error parsing page %s: %v", path, err)
			}
			pagesByName[page.Name] = page
			for _, tag := range page.Header.Tags {
				tagNames[uri.GetTagPath(tag)] = tag
			}
		} else if strings.HasSuffix(path, ".tmpl") {
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

	tocByLanguage, err := indexPages(pagesByName, translationsByName)
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

type filePopulator func(w io.Writer) error

// Write converts the site contents to HTML and writes it to disk.
func (c *Contents) Write() error {
	t := &templates.Templates{
		TemplatePath: c.TemplatePath,
		WebRoot:      c.WebRoot,
	}
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
		err := makeFile(
			filepath.Join(c.OutputPath, page.Name+".html"),
			func(w io.Writer) error {
				return c.outputPage(w, t, page)
			})
		if err != nil {
			return fmt.Errorf("Error rendering page %s: %v", page.Name, err)
		}
	}
	for lang, languageToc := range c.Toc {
		years := getYearsInReverseOrder(languageToc.All.ByYear)
		for year, toc := range languageToc.All.ByYear {
			err := makeFile(
				fmt.Sprintf("%s-%s-%d.html", filepath.Join(c.OutputPath, "toc", "toc"), lang, year),
				func(w io.Writer) error {
					return c.outputToc(w, t, lang, toc, year, years, languageToc.All.TotalCount, "", uri.Concat(c.WebRoot, "toc", "toc-"+lang))
				})
			if err != nil {
				return err
			}
		}
		err := makeFile(
			fmt.Sprintf("%s-%s.html", filepath.Join(c.OutputPath, "toc", "toc"), lang),
			func(w io.Writer) error {
				return c.outputIndex(w, t, lang, years, languageToc.All, "", uri.Concat(c.WebRoot, "toc", "toc-"))
			})
		if err != nil {
			return err
		}
		for tag, tagToc := range languageToc.ByTag {
			years := getYearsInReverseOrder(tagToc.ByYear)
			for year, toc := range tagToc.ByYear {
				err := makeFile(
					fmt.Sprintf("%s-%s-%d.html", filepath.Join(c.OutputPath, "tags", tag), lang, year),
					func(w io.Writer) error {
						return c.outputToc(w, t, lang, toc, year, years, tagToc.TotalCount, c.Tags[tag], uri.Concat(c.WebRoot, "tags", tag+"-"+lang))
					})
				if err != nil {
					return err
				}
			}
			err := makeFile(
				fmt.Sprintf("%s-%s-%s.html", filepath.Join(c.OutputPath, "tags", "toc"), tag, lang),
				func(w io.Writer) error {
					return c.outputIndex(w, t, lang, years, tagToc, c.Tags[tag], uri.Concat(c.WebRoot, "tags", "toc-"+tag))
				})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Contents) copyFile(name string) error {
	inName := filepath.Join(c.InputPath, name)
	outName := filepath.Join(c.OutputPath, name)
	inFile, err := os.Open(inName)
	if err != nil {
		return err
	}
	err = makeFile(outName, func(w io.Writer) error {
		_, err := io.Copy(w, inFile)
		return err
	})
	err2 := inFile.Close()
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	stat, err := os.Stat(inName)
	if err != nil {
		return err
	}
	return os.Chtimes(outName, stat.ModTime(), stat.ModTime())
}

func makeFile(name string, populator filePopulator) error {
	err := os.MkdirAll(filepath.Dir(name), 0o755)
	if err != nil {
		return err
	}
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	err = populator(file)
	return err
}
