package renderer

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"time"

	"jtweb/page"
	"jtweb/renderer/extensions"

	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

var markdown goldmark.Markdown = goldmark.New(
	goldmark.WithRendererOptions(
		html.WithUnsafe(),
	),
	goldmark.WithExtensions(
		extensions.HeaderExtension,
		extensions.YouTubeExtension,
		extension.GFM,
		extension.Typographer,
		highlighting.NewHighlighting(highlighting.WithStyle("igor")),
		mathjax.MathJax,
	),
)

var sanitizer = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").Matching(bluemonday.SpaceSeparatedTokens).OnElements("div", "span")
	p.AllowAttrs("class", "name").Matching(bluemonday.SpaceSeparatedTokens).OnElements("a")
	p.AllowAttrs("rel").Matching(regexp.MustCompile(`^nofollow$`)).OnElements("a")
	p.AllowAttrs("aria-hidden").Matching(regexp.MustCompile(`^true$`)).OnElements("a")
	p.AllowAttrs("type").Matching(regexp.MustCompile(`^checkbox$`)).OnElements("input")
	p.AllowAttrs("checked", "disabled").Matching(regexp.MustCompile(`^$`)).OnElements("input")
	p.AllowDataURIImages()
	return p
}()

// Parse reads a page in Markdown format and parses it.
func Parse(r io.Reader, name string) (*page.Page, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	src := buf.Bytes()
	md := markdown.Parser().Parse(text.NewReader(src))
	header, err := findHeader(md, src)
	if err != nil {
		return nil, err
	}
	page := &page.Page{Name: name, Source: src, Root: md, Header: header}
	return page, nil
}

// Render renders the page in HTML format.
func Render(w io.Writer, p *page.Page) error {
	buf := &bytes.Buffer{}
	err := markdown.Renderer().Render(buf, p.Source, p.Root)
	if err != nil {
		return err
	}
	buf.WriteTo(w)
	_, err = sanitizer.SanitizeReader(buf).WriteTo(w)
	return err
}

func findHeader(root ast.Node, src []byte) (page.HeaderData, error) {
	var header page.HeaderData
	err := ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if n.Kind() == extensions.KindHeaderBlock {
			rawHeader, err := n.(*extensions.HeaderBlock).ParseContents(src)
			if err != nil {
				return ast.WalkStop, err
			}
			header, err = parseHeader(rawHeader)
			if err != nil {
				return ast.WalkStop, err
			}
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return header, err
}

func parseHeader(rawHeader *extensions.RawHeader) (page.HeaderData, error) {
	var out page.HeaderData

	if rawHeader.Title == "" {
		return out, fmt.Errorf("Missing title")
	}

	out.Title = rawHeader.Title
	out.Language = rawHeader.Language
	if out.Language == "" {
		out.Language = "en"
	}
	out.Summary = rawHeader.Summary
	if rawHeader.PublishDate != "" {
		d, err := parseDate(rawHeader.PublishDate)
		if err == nil {
			out.PublishDate = d
		}
	}
	out.HidePublishDate = rawHeader.HidePublishDate
	out.AuthorName = rawHeader.AuthorName
	out.AuthorURI = rawHeader.AuthorURI
	out.HideAuthor = rawHeader.HideAuthor
	out.Tags = rawHeader.Tags
	out.NoIndex = rawHeader.NoIndex
	out.OldURI = rawHeader.OldURI
	out.TranslationOf = rawHeader.TranslationOf
	return out, nil
}

var dateFormats []string = []string{"2006-01-02", "01/02/2006", "02-01-2006"}
var timeFormats []string = []string{"3:04:05pm", "3:04pm", "15:04:05", "15:04"}

func parseDate(str string) (time.Time, error) {
	d, err := time.Parse("2006-01-02T15:04:05-0700", str)
	if err == nil {
		return d, nil
	}
	for _, df := range dateFormats {
		for _, tf := range timeFormats {
			d, err = time.Parse(df+" "+tf+" -0700 MST", str)
			if err == nil {
				return d, nil
			}
			d, err = time.Parse(df+" "+tf+" -0700", str)
			if err == nil {
				return d, nil
			}
			d, err = time.Parse(df+" "+tf, str)
			if err == nil {
				return d, nil
			}
		}
		d, err = time.Parse(df, str)
		if err == nil {
			return d, nil
		}
	}
	return d, err
}
