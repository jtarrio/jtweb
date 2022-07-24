package site

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"jacobo.tarrio.org/jtweb/page"
	"jacobo.tarrio.org/jtweb/uri"
)

const redirectPlaceholder = "### REDIRECTS ###"

func (c *Contents) outputHtaccess(name string) error {
	inName := filepath.Join(c.InputPath, name)
	outName := filepath.Join(c.OutputPath, name)
	inFile, err := os.Open(inName)
	if err != nil {
		return err
	}
	err = makeFile(outName, func(w io.Writer) error {
		var eof bool
		br := bufio.NewReader(inFile)
		bw := bufio.NewWriter(w)
		for {
			line, err := br.ReadString('\n')
			if err == io.EOF {
				err = nil
				eof = true
			}
			if err != nil {
				return err
			}
			if strings.TrimSpace(line) == redirectPlaceholder {
				err = c.writeRedirects(bw)
				if err != nil {
					return err
				}
			} else {
				_, err = bw.WriteString(line)
				if err != nil {
					return err
				}
			}
			if eof {
				bw.Flush()
				return nil
			}
		}
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

	parsed, err := url.Parse(uri.Concat(c.GetWebRoot(p.Header.Language), p.Name+".html"))
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
