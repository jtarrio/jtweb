package site

import (
	"bufio"
	"fmt"
	goio "io"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"jacobo.tarrio.org/jtweb/page"
	"jacobo.tarrio.org/jtweb/uri"
)

const redirectPlaceholder = "### REDIRECTS ###"

func (c *Contents) outputHtaccess(name string) error {
	source := c.Config.Files().Input().GoTo(name)
	target := c.Config.Generator().Output().GoTo(name)
	input, err := source.Read()
	if err != nil {
		return err
	}
	output, err := target.Create()
	if err != nil {
		input.Close()
		return err
	}
	eof := false
	br := bufio.NewReader(input)
	bw := bufio.NewWriter(output)
	for !eof {
		line, err := br.ReadString('\n')
		if err == goio.EOF {
			err = nil
			eof = true
		}
		if err != nil {
			break
		}
		if strings.TrimSpace(line) == redirectPlaceholder {
			err = c.writeRedirects(bw)
			if err != nil {
				break
			}
		} else {
			_, err = bw.WriteString(line)
			if err != nil {
				break
			}
		}
		if eof {
			bw.Flush()
			err = nil
		}
	}
	input.Close()
	output.Close()
	if err != nil {
		return err
	}
	stat, err := source.Stat()
	if err != nil {
		return err
	}
	return target.Chtime(stat.ModTime)
}

type redirectPattern struct {
	oldPath     string
	newPath     string
	publishDate time.Time
}

func (c *Contents) writeRedirects(bw *bufio.Writer) error {
	var pats []redirectPattern
	for _, page := range c.Pages {
		patterns, err := c.makeRedirectPatterns(page)
		if err != nil {
			return err
		}
		pats = append(pats, patterns...)
	}
	sort.SliceStable(pats, func(i, j int) bool {
		a := pats[i]
		b := pats[j]
		if !a.publishDate.Equal(b.publishDate) {
			return a.publishDate.After(b.publishDate)
		}
		if a.newPath != b.newPath {
			return a.newPath < b.newPath
		}
		return a.oldPath < b.oldPath
	})

	for _, pat := range pats {
		_, err := bw.WriteString(fmt.Sprintf("RewriteRule ^%s$ %s [R=301,L]\n", pat.oldPath, pat.newPath))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Contents) makeRedirectPatterns(p *page.Page) ([]redirectPattern, error) {
	out := make([]redirectPattern, 0)

	parsed, err := url.Parse(uri.Concat(c.Config.Site(p.Header.Language).WebRoot(), string(p.Name)+".html"))
	if err != nil {
		return nil, err
	}
	newPath := parsed.EscapedPath()
	publishDate := p.Header.PublishDate
	if p.Header.HidePublishDate {
		publishDate = time.Time{}
	}

	for _, oldURI := range p.Header.OldURI {
		parsed, err := url.Parse(oldURI)
		if err != nil {
			return nil, err
		}
		path := parsed.EscapedPath()
		if path[0] == '/' {
			path = path[1:]
		}
		out = append(out, redirectPattern{
			oldPath:     regexp.QuoteMeta(path),
			newPath:     newPath,
			publishDate: publishDate})
	}

	return out, nil
}
