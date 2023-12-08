package site

import (
	"fmt"
	goio "io"
	"path/filepath"
	"sort"
	"strings"

	"jacobo.tarrio.org/jtweb/config"
	"jacobo.tarrio.org/jtweb/io"
	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/page"
	"jacobo.tarrio.org/jtweb/uri"
)

// TagId is a custom type for a tag's identifier.
type TagId string

// Contents contains the parsed and indexed content of the site.
type Contents struct {
	Config       config.Config
	Files        []string
	Templates    []string
	Pages        map[page.Name]*page.Page
	Tags         map[TagId]string
	Toc          GlobalTableOfContents
	Translations map[page.Name][]Translation
}

// GlobalTableOfContents contains the tables of contents for every language.
type GlobalTableOfContents map[languages.Language]LanguageTableOfContents

// LanguageTableOfContents contains the tables of contents for every year in a language.
type LanguageTableOfContents struct {
	// TOC for all pages.
	All TableOfContents
	// TOC for each tag.
	ByTag map[TagId]TableOfContents
	// Map from each page name to the immediately newer page's name.
	NewerPages map[page.Name]page.Name
	// Map from each page name to the immediately older page's name.
	OlderPages map[page.Name]page.Name
}

// TableOfContents contains a list of pages.
type TableOfContents []page.Name

// Translation contains information about a page translation.
type Translation struct {
	Name     page.Name
	Language languages.Language
}

// Read parses the whole site contents.
func Read(s config.Config) (*Contents, error) {
	files := make([]string, 0)
	templates := make([]string, 0)
	pagesByName := make(map[page.Name]*page.Page)
	tagIds := make(map[TagId]string)
	err := s.Files().Input().ForAllFiles(func(file io.File, err error) error {
		if err != nil {
			return err
		}
		name := file.Name()
		if strings.HasSuffix(name, ".md") {
			page, err := parsePage(file)
			if err != nil {
				return fmt.Errorf("error parsing page %s: %v", file.Name(), err)
			}
			if page.Header.PublishDate.After(s.Generator().PublishUntil()) {
				return nil
			}
			pagesByName[page.Name] = page
			for _, tag := range page.Header.Tags {
				tagIds[tagId(tag)] = tag
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

	tocByLanguage, err := indexPages(pagesByName, translationsByName, s.Generator().HideUntranslated())
	if err != nil {
		return nil, err
	}

	c := Contents{
		Config:       s,
		Files:        files,
		Templates:    templates,
		Pages:        pagesByName,
		Tags:         tagIds,
		Toc:          tocByLanguage,
		Translations: translationsByName,
	}
	return &c, nil
}

// tagId returns a TagId for the given tag.
func tagId(tag string) TagId {
	return TagId(uri.GetTagPath(tag))
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
		err := c.makeFile(
			string(page.Name)+".html",
			func(w io.Output) error {
				return c.OutputAsPage(w, page)
			})
		if err != nil {
			return fmt.Errorf("error rendering page %s: %w", page.Name, err)
		}
	}
	for lang, languageToc := range c.Toc {
		err := c.makeFile(
			fmt.Sprintf("toc/toc-%s.html", lang.Code()),
			func(w io.Output) error {
				return c.outputToc(w, lang, languageToc.All, "")
			})
		if err != nil {
			return err
		}
		for tag, tagToc := range languageToc.ByTag {
			err := c.makeFile(
				fmt.Sprintf("tags/%s-%s.html", tag, lang.Code()),
				func(w io.Output) error {
					return c.outputToc(w, lang, tagToc, c.Tags[tag])
				})
			if err != nil {
				return err
			}
		}
		err = c.makeFile(
			fmt.Sprintf("rss/%s.xml", lang.Code()),
			func(w io.Output) error {
				return c.outputRss(w, lang)
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Contents) copyFile(name string) error {
	in := c.Config.Files().Input().GoTo(name)
	out := c.Config.Generator().Output().GoTo(name)
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
	file := c.Config.Generator().Output().GoTo(path)
	output, err := file.Create()
	if err != nil {
		return err
	}
	defer output.Close()
	err = populator(output)
	return err
}

func getTranslationsByName(pages map[page.Name]*page.Page) (map[page.Name][]Translation, error) {
	translationsByName := make(map[page.Name]map[languages.Language]page.Name)
	for _, pg := range pages {
		if pg.Header.Draft {
			continue
		}
		translations := translationsByName[pg.Name]
		if translations == nil {
			translations = make(map[languages.Language]page.Name)
			translationsByName[pg.Name] = translations
		}
		if translations[pg.Header.Language] != "" && translations[pg.Header.Language] != pg.Name {
			return nil, fmt.Errorf(
				"two pages claim language [%s] for the same content: [%s] and [%s]",
				pg.Header.Language,
				translations[pg.Header.Language],
				pg.Name,
			)
		}
		translations[pg.Header.Language] = pg.Name
		if pg.Header.TranslationOf != "" {
			otherTranslations := translationsByName[pg.Header.TranslationOf]
			for l, p := range otherTranslations {
				if translations[l] != "" && translations[l] != p {
					return nil, fmt.Errorf(
						"translation of [%s] to language [%s] has two conflicting values: [%s] and [%s]",
						pg.Name,
						l,
						translations[l],
						p,
					)
				}
				translations[l] = p
			}
			translationsByName[pg.Header.TranslationOf] = translations
		}
	}
	translations := make(map[page.Name][]Translation)
	for name, trans := range translationsByName {
		for code, page := range trans {
			if code != pages[name].Header.Language {
				translations[name] = append(translations[name], Translation{Language: code, Name: page})
			}
		}
	}
	return translations, nil
}

func indexPages(pages map[page.Name]*page.Page, translationsByName map[page.Name][]Translation, hideUntranslated bool) (GlobalTableOfContents, error) {
	langs := make(map[languages.Language]bool)
	for _, page := range pages {
		langs[page.Header.Language] = true
	}
	toc := make(GlobalTableOfContents)
	for lang := range langs {
		languageToc := LanguageTableOfContents{
			NewerPages: make(map[page.Name]page.Name),
			OlderPages: make(map[page.Name]page.Name),
		}
		allNames := makeNameIndex(pages, lang, translationsByName, hideUntranslated)
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
		languageToc.ByTag = make(map[TagId]TableOfContents)
		for tag, allNamesOfTag := range allNamesByTag {
			languageToc.ByTag[tag] = allNamesOfTag
		}
		toc[lang] = languageToc
	}

	return toc, nil
}

func makeNameIndex(pages map[page.Name]*page.Page, language languages.Language, translationsByName map[page.Name][]Translation, hideUntranslated bool) []page.Name {
	allNames := make([]page.Name, 0, len(pages))
	for name, page := range pages {
		if page.Header.NoIndex || page.Header.Draft {
			continue
		}
		if page.Header.Language == language {
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
		translations := make(map[languages.Language]bool)
		translations[page.Header.Language] = true
		for _, tl := range translationsByName[name] {
			translations[tl.Language] = true
		}
		if translations[language] {
			continue
		}
		// We want to show only the preferred translation.
		wanted := make([]languages.Language, 0, len(translations))
		for lang := range translations {
			wanted = append(wanted, lang)
		}
		sort.Sort(languages.LanguageSlice(wanted))
		preferred := language.PreferredLanguage(wanted)
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

func groupByTag(names []page.Name, pages map[page.Name]*page.Page) map[TagId]TableOfContents {
	byTag := make(map[TagId]TableOfContents)
	for _, name := range names {
		for _, tag := range pages[name].Header.Tags {
			id := tagId(tag)
			byTag[id] = append(byTag[id], name)
		}
	}
	return byTag
}
