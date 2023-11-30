package site

import (
	"io"
	"strings"

	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/page"
	"jacobo.tarrio.org/jtweb/uri"

	"github.com/gorilla/feeds"
)

func (c *Contents) outputRss(w io.Writer, lang languages.Language) error {
	toc := c.Toc[lang]
	pages := make([]*page.Page, 0, 5)
	if len(toc.All) > 0 {
		pageName := toc.All[0]
		for i := 0; i < cap(pages) && pageName != ""; i++ {
			pages = append(pages, c.Pages[pageName])
			pageName = toc.OlderPages[pageName]
		}
	}

	feed := &feeds.Feed{
		Title: c.Config.GetSiteName(lang),
		Link:  &feeds.Link{Href: c.Config.GetSiteURI(lang)},
	}
	feed.Items = make([]*feeds.Item, len(pages))
	for i, p := range pages {
		var sb strings.Builder
		err := p.Render(&sb)
		if err != nil {
			return err
		}
		feed.Items[i] = &feeds.Item{
			Title:       p.Header.Title,
			Link:        &feeds.Link{Href: uri.Concat(c.Config.GetWebRoot(lang), string(p.Name)) + ".html"},
			Author:      &feeds.Author{Name: p.Header.AuthorName},
			Created:     p.Header.PublishDate,
			Description: sb.String(),
		}
	}

	rss, err := feed.ToRss()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(rss))
	return err
}
