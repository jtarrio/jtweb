package site

import (
	"bufio"
	"fmt"
	"io"
	"jtweb/page"
	"jtweb/uri"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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

func (c *Contents) writeRedirects(bw *bufio.Writer) error {
	var pats patternList
	for _, page := range c.Pages {
		patterns, err := c.makeRedirectPatterns(page)
		if err != nil {
			return err
		}
		pats = append(pats, patterns...)
	}
	sort.Stable(pats)

	for _, pat := range pats {
		err := pat.Write(bw)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Contents) makeRedirectPatterns(p *page.Page) ([]redirectPattern, error) {
	out := make([]redirectPattern, 0)

	parsed, err := url.Parse(uri.Concat(c.WebRoot, p.Name+".html"))
	if err != nil {
		return nil, err
	}
	newPath := parsed.EscapedPath()

	for _, oldURI := range p.Header.OldURI {
		parsed, err := url.Parse(oldURI)
		if err != nil {
			return nil, err
		}
		path := parsed.EscapedPath()
		if path[0] == '/' {
			path = path[1:]
		}
		if strings.HasPrefix(path, "node/") {
			out = append(out, &uriBasedPattern{oldPath: "(es/|gl/)?" + regexp.QuoteMeta(path), newPath: newPath})
		} else {
			out = append(out, &uriBasedPattern{oldPath: regexp.QuoteMeta(path), newPath: newPath})
		}
	}

	return out, nil
}

type redirectPattern interface {
	Write(bw *bufio.Writer) error
}

type uriBasedPattern struct {
	oldPath string
	newPath string
}

func (p *uriBasedPattern) Write(bw *bufio.Writer) error {
	_, err := bw.WriteString(fmt.Sprintf("RewriteRule\t^%s$\t%s\t[R,L]\n", p.oldPath, p.newPath))
	return err
}

type patternList []redirectPattern

func (p patternList) Len() int {
	return len(p)
}

func (p patternList) Less(i, j int) bool {
	a := p[i].(*uriBasedPattern)
	b := p[j].(*uriBasedPattern)
	return a.oldPath < b.oldPath || (a.oldPath == b.oldPath && a.newPath < b.newPath)
}

func (p patternList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
