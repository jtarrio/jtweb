package service

import (
	"strings"

	gohtml "html"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"jacobo.tarrio.org/jtweb/comments"
)

type Renderer interface {
	Render(src comments.Markdown) (comments.Html, error)
}

type escapeRenderer struct{}

func NewEscapeRenderer() Renderer {
	return &escapeRenderer{}
}

func (r *escapeRenderer) Render(src comments.Markdown) (comments.Html, error) {
	return comments.Html(gohtml.EscapeString(string(src))), nil
}

type markdownRenderer struct {
	markdown  goldmark.Markdown
	sanitizer *bluemonday.Policy
}

func NewMarkdownRenderer() Renderer {
	return &markdownRenderer{
		markdown: goldmark.New(
			goldmark.WithRendererOptions(
				html.WithUnsafe(),
			),
			goldmark.WithExtensions(
				extension.GFM,
				extension.Typographer,
			),
		),
		sanitizer: commentPolicy()}
}

func commentPolicy() *bluemonday.Policy {
	p := bluemonday.StrictPolicy()
	p.AllowStandardAttributes()
	p.AllowStandardURLs()
	p.RequireParseableURLs(true)
	p.AllowElements("br", "div", "hr", "p", "span", "wbr")
	p.AllowAttrs("href").OnElements("a")
	p.RequireNoFollowOnLinks(true)
	p.AllowElements("abbr", "acronym", "cite", "code", "dfn", "em",
		"figcaption", "mark", "s", "samp", "strong", "sub", "sup", "var")
	p.AllowElements("b", "i", "pre", "small", "strike", "tt", "u")
	p.AllowLists()
	return p
}

func (r *markdownRenderer) Render(src comments.Markdown) (comments.Html, error) {
	bytes := []byte(src)
	root := r.markdown.Parser().Parse(text.NewReader(bytes))
	sb := strings.Builder{}
	err := r.markdown.Renderer().Render(&sb, bytes, root)
	if err != nil {
		return "", err
	}

	html := comments.Html(r.sanitizer.Sanitize(sb.String()))
	return html, nil
}
