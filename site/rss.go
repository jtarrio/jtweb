package site

import (
	"io"
	"jtweb/page"
	"jtweb/renderer"
	"jtweb/renderer/templates"
	"jtweb/uri"
	"strings"

	"github.com/gorilla/feeds"
)

func (c *Contents) outputRss(w io.Writer, t *templates.Templates, lang string) error {
	toc := c.Toc[lang]
	pageName := toc.All[0]
	pages := make([]*page.Page, 0, 5)
	for i := 0; i < cap(pages) && pageName != ""; i++ {
		pages = append(pages, c.Pages[pageName])
		pageName = toc.OlderPages[pageName]
	}

	feed := &feeds.Feed{
		Title: c.Config.SiteName,
		Link:  &feeds.Link{Href: c.Config.SiteURI},
	}
	feed.Items = make([]*feeds.Item, len(pages))
	for i, p := range pages {
		var sb strings.Builder
		err := renderer.Render(&sb, p)
		if err != nil {
			return err
		}
		feed.Items[i] = &feeds.Item{
			Title:       p.Header.Title,
			Link:        &feeds.Link{Href: uri.Concat(c.Config.WebRoot, p.Name) + ".html"},
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
