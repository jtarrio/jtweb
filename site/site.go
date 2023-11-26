package site

import (
	"fmt"
	goio "io"
	"path/filepath"
	"sort"
	"strings"

	"jacobo.tarrio.org/jtweb/languages"
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
				return c.OutputAsPage(w, t, page)
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

func getTranslationsByName(pages map[string]*page.Page) (map[string][]Translation, error) {
	translationsByName := make(map[string]map[string]string)
	for _, page := range pages {
		if page.Header.Draft {
			continue
		}
		translations := translationsByName[page.Name]
		if translations == nil {
			translations = make(map[string]string)
			translationsByName[page.Name] = translations
		}
		if translations[page.Header.Language] != "" && translations[page.Header.Language] != page.Name {
			return nil, fmt.Errorf(
				"two pages claim language [%s] for the same content: [%s] and [%s]",
				page.Header.Language,
				translations[page.Header.Language],
				page.Name,
			)
		}
		translations[page.Header.Language] = page.Name
		if page.Header.TranslationOf != "" {
			otherTranslations := translationsByName[page.Header.TranslationOf]
			for l, p := range otherTranslations {
				if translations[l] != "" && translations[l] != p {
					return nil, fmt.Errorf(
						"translation of [%s] to language [%s] has two conflicting values: [%s] and [%s]",
						page.Name,
						l,
						translations[l],
						p,
					)
				}
				translations[l] = p
			}
			translationsByName[page.Header.TranslationOf] = translations
		}
	}
	translations := make(map[string][]Translation)
	for name, trans := range translationsByName {
		for code, page := range trans {
			if code != pages[name].Header.Language {
				translations[name] = append(translations[name], Translation{Language: code, Name: page})
			}
		}
	}
	return translations, nil
}

func indexPages(pages map[string]*page.Page, translationsByName map[string][]Translation, hideUntranslated bool) (GlobalTableOfContents, error) {
	langs := make(map[string]bool)
	for _, page := range pages {
		langs[page.Header.Language] = true
	}
	toc := make(GlobalTableOfContents)
	for lang := range langs {
		languageToc := LanguageTableOfContents{
			NewerPages: make(map[string]string),
			OlderPages: make(map[string]string),
		}
		language := languages.FindByCodeWithFallback(lang, languages.LanguageEn)
		allNames := makeNameIndex(pages, language, translationsByName, hideUntranslated)
		for i, name := range allNames {
			if i > 0 {
				languageToc.NewerPages[name] = allNames[i-1]
			}
			if i < len(allNames)-1 {
				languageToc.OlderPages[name] = allNames[i+1]
			}
		}
		languageToc.All = allNames
		allNamesByTag := groupByTag(allNames, pages)
		languageToc.ByTag = make(map[string]TableOfContents)
		for tag, allNamesOfTag := range allNamesByTag {
			languageToc.ByTag[tag] = allNamesOfTag
		}
		toc[lang] = languageToc
	}

	return toc, nil
}

func makeNameIndex(pages map[string]*page.Page, language languages.Language, translationsByName map[string][]Translation, hideUntranslated bool) []string {
	allNames := make([]string, 0, len(pages))
	for name, page := range pages {
		if page.Header.NoIndex || page.Header.Draft {
			continue
		}
		if page.Header.Language == language.Code() {
			// If the page is in the index language, add it.
			allNames = append(allNames, name)
			continue
		}
		// Skip if we are hiding untranslated
		if hideUntranslated {
			continue
		}
		// If it has no translations, add it.
		if translationsByName[name] == nil {
			allNames = append(allNames, name)
			continue
		}
		// If the index language is among the translations, skip it.
		translations := make(map[string]bool)
		translations[page.Header.Language] = true
		for _, tl := range translationsByName[name] {
			translations[tl.Language] = true
		}
		if translations[language.Code()] {
			continue
		}
		// We want to show only the preferred translation.
		wanted := make([]string, 0, len(translations))
		for lang := range translations {
			wanted = append(wanted, lang)
		}
		sort.Strings(wanted)
		preferred := language.PreferredLanguage(wanted)
		if preferred == "" {
			preferred = wanted[0]
		}
		if preferred == page.Header.Language {
			allNames = append(allNames, name)
		}
	}
	sort.SliceStable(allNames, func(i, j int) bool {
		a := pages[allNames[i]]
		b := pages[allNames[j]]
		if a.Header.HidePublishDate {
			return b.Header.HidePublishDate && a.Name < b.Name
		}
		if b.Header.HidePublishDate {
			return true
		}
		return a.Header.PublishDate.After(b.Header.PublishDate)
	})
	return allNames
}

func groupByTag(names []string, pages map[string]*page.Page) map[string]TableOfContents {
	byTag := make(map[string]TableOfContents)
	for _, name := range names {
		for _, tag := range pages[name].Header.Tags {
			tagPath := uri.GetTagPath(tag)
			byTag[tagPath] = append(byTag[tagPath], name)
		}
	}
	return byTag
}
