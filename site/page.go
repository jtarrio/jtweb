package site

import (
	"fmt"
	"html/template"
	"io"
	"jtweb/languages"
	"jtweb/page"
	"jtweb/renderer"
	"jtweb/renderer/templates"
	"jtweb/uri"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func parsePage(path string, name string) (*page.Page, error) {
	name = filepath.ToSlash(name[:len(name)-3])
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return renderer.Parse(file, name)
}

func getTranslationsByName(pages map[string]*page.Page) (map[string][]Translation, error) {
	translationsByName := make(map[string]map[string]string)
	for _, page := range pages {
		translations := translationsByName[page.Name]
		if translations == nil {
			translations = make(map[string]string)
			translationsByName[page.Name] = translations
		}
		if translations[page.Header.Language] != "" && translations[page.Header.Language] != page.Name {
			return nil, fmt.Errorf(
				"Two pages claim language [%s] for the same content: [%s] and [%s]",
				page.Header.Language,
				translations[page.Header.Language],
				page.Name,
			)
		}
		translations[page.Header.Language] = page.Name
		if page.Header.TranslationOf != "" {
			otherTranslations := translationsByName[page.Header.TranslationOf]
			if otherTranslations != nil {
				for l, p := range otherTranslations {
					if translations[l] != "" && translations[l] != p {
						return nil, fmt.Errorf(
							"Translation of [%s] to language [%s] has two conflicting values: [%s] and [%s]",
							page.Name,
							l,
							translations[l],
							p,
						)
					}
					translations[l] = p
				}
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

func indexPages(pages map[string]*page.Page, translationsByName map[string][]Translation) (GlobalTableOfContents, error) {
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
		allNames := makeNameIndex(pages, language, translationsByName)
		for i, name := range allNames {
			if i > 0 {
				languageToc.NewerPages[name] = allNames[i-1]
			}
			if i < len(allNames)-1 {
				languageToc.OlderPages[name] = allNames[i+1]
			}
		}
		languageToc.All = TableOfContents{
			TotalCount: len(allNames),
			ByYear:     groupByPublishYear(allNames, pages),
		}
		allNamesByTag := groupByTag(allNames, pages)
		languageToc.ByTag = make(map[string]TableOfContents)
		for tag, allNamesOfTag := range allNamesByTag {
			languageToc.ByTag[tag] = TableOfContents{
				TotalCount: len(allNamesOfTag),
				ByYear:     groupByPublishYear(allNamesOfTag, pages),
			}
		}
		toc[lang] = languageToc
	}

	return toc, nil
}

func makeNameIndex(pages map[string]*page.Page, language languages.Language, translationsByName map[string][]Translation) []string {
	allNames := make([]string, 0, len(pages))
	for name, page := range pages {
		if page.Header.NoIndex {
			continue
		}
		if page.Header.Language == language.Code() || translationsByName[name] == nil {
			// If the page is in the index language or has no translations, add it.
			allNames = append(allNames, name)
		} else {
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

func groupByPublishYear(nameList []string, pages map[string]*page.Page) map[int]SingleTableOfContents {
	groups := make(map[int]SingleTableOfContents)
	for _, name := range nameList {
		header := pages[name].Header
		year := header.PublishDate.Year()
		if header.HidePublishDate {
			year = 0
		}
		groups[year] = append(groups[year], name)
	}
	return groups
}

func groupByTag(names []string, pages map[string]*page.Page) map[string]SingleTableOfContents {
	byTag := make(map[string]SingleTableOfContents)
	for _, name := range names {
		for _, tag := range pages[name].Header.Tags {
			tagPath := uri.GetTagPath(tag)
			byTag[tagPath] = append(byTag[tagPath], name)
		}
	}
	return byTag
}

func getYearsInReverseOrder(m map[int]SingleTableOfContents) []int {
	years := make([]int, len(m))
	i := 0
	for k := range m {
		years[i] = k
		i++
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))
	return years
}

func (c *Contents) outputPage(w io.Writer, t *templates.Templates, page *page.Page) error {
	tmpl, err := t.GetPageTemplate(page.Header.Language)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, c.makePageData(page))
}

func (c *Contents) outputToc(w io.Writer, t *templates.Templates, lang string, names []string, year int, years []int, totalCount int, tag string, baseURI string) error {
	tmpl, err := t.GetTocTemplate(lang)
	if err != nil {
		return err
	}
	undatedStories := false
	storyYears := make([]int, 0, len(names))
	for _, y := range years {
		if y == 0 {
			undatedStories = true
		} else if y != year {
			storyYears = append(storyYears, y)
		}
	}

	stories := make([]templates.PageData, len(names))
	for i, name := range names {
		stories[i] = c.makePageData(c.Pages[name])
	}

	tocData := templates.TocData{
		BaseURI:        baseURI,
		Tag:            tag,
		Year:           year,
		YearCount:      len(names),
		TotalCount:     totalCount,
		Stories:        stories,
		StoryYears:     storyYears,
		UndatedStories: undatedStories,
	}

	return tmpl.Execute(w, tocData)
}

func (c *Contents) makePageData(page *page.Page) templates.PageData {
	sb := strings.Builder{}
	renderer.Render(&sb, page)

	pageData := templates.PageData{
		Title:     page.Header.Title,
		Permalink: c.makePageURI(page),
		Author:    templates.LinkData{},
		Summary:   page.Header.Summary,
		Tags:      page.Header.Tags,
		Content:   template.HTML(sb.String()),
	}
	if !page.Header.HidePublishDate {
		pageData.PublishDate = page.Header.PublishDate
	}
	if !page.Header.HideAuthor {
		pageData.Author = templates.LinkData{
			Name: page.Header.AuthorName,
			URI:  page.Header.AuthorURI,
		}
		if pageData.Author.Name == "" && pageData.Author.URI == "" {
			pageData.Author.Name = c.Config.AuthorName
			pageData.Author.URI = c.Config.AuthorURI
		}
	}
	newer := c.Toc[page.Header.Language].NewerPages[page.Name]
	if newer != "" {
		pageData.NewerPage = templates.LinkData{
			URI:  c.makePageURI(c.Pages[newer]),
			Name: c.Pages[newer].Header.Title,
		}
	}
	older := c.Toc[page.Header.Language].OlderPages[page.Name]
	if older != "" {
		pageData.OlderPage = templates.LinkData{
			URI:  c.makePageURI(c.Pages[older]),
			Name: c.Pages[older].Header.Title,
		}
	}
	translations := c.Translations[page.Name]
	if translations != nil {
		for _, t := range translations {
			translation := c.Pages[t.Name]
			pageData.Translations = append(
				pageData.Translations,
				templates.TranslationData{
					Name:     translation.Header.Title,
					URI:      c.makePageURI(translation),
					Language: t.Language,
				})
		}
	}
	return pageData
}

func (c *Contents) makePageURI(p *page.Page) string {
	return uri.Concat(c.WebRoot, p.Name+".html")
}

func (c *Contents) outputIndex(w io.Writer, t *templates.Templates, lang string, years []int, toc TableOfContents, tag string, uriBase string) error {
	tmpl, err := t.GetIndexTocTemplate(lang)
	if err != nil {
		return err
	}
	yearData := make([]templates.YearData, len(years))
	for i, year := range years {
		yearData[i] = templates.YearData{
			Year:  year,
			Count: len(toc.ByYear[year]),
			Tags:  c.getTags(toc.ByYear[year]),
		}
	}
	indexTocData := templates.IndexTocData{
		BaseURI:    uriBase,
		Tag:        tag,
		TotalCount: toc.TotalCount,
		Years:      yearData,
	}
	return tmpl.Execute(w, indexTocData)
}

func (c *Contents) getTags(stories []string) []string {
	tagMap := make(map[string]string)
	for _, storyName := range stories {
		for _, tag := range c.Pages[storyName].Header.Tags {
			tagMap[uri.GetTagPath(tag)] = tag
		}
	}
	tags := make([]string, 0, len(tagMap))
	for _, tag := range tagMap {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}
