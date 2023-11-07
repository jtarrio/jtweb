package renderer

import (
	"bytes"
	"io"
	"regexp"

	"jacobo.tarrio.org/jtweb/page"
	"jacobo.tarrio.org/jtweb/renderer/extensions"

	htmlformatter "github.com/alecthomas/chroma/formatters/html"
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
		extensions.MultipleImageExtension,
		extensions.ImageCaptionExtension,
		extension.GFM,
		extension.Typographer,
		highlighting.NewHighlighting(
			highlighting.WithStyle("igor"),
			highlighting.WithFormatOptions(htmlformatter.TabWidth(4))),
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
			var err error
			header, err = n.(*extensions.HeaderBlock).ParseContents(src)
			return ast.WalkStop, err
		}
		return ast.WalkContinue, nil
	})
	return header, err
}
