package renderer

import (
	"bytes"
	"io"

	"jacobo.tarrio.org/jtweb/renderer/extensions"

	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

var markdown goldmark.Markdown = goldmark.New(
	goldmark.WithParserOptions(parser.WithAttribute()),
	goldmark.WithRendererOptions(
		html.WithUnsafe(),
	),
	goldmark.WithExtensions(
		extensions.HeaderExtension,
		extensions.YouTubeExtension,
		extensions.MultipleImageExtension,
		extensions.ImageCaptionExtension,
		extension.Footnote,
		extension.GFM,
		extension.Typographer,
		highlighting.NewHighlighting(highlighting.WithStyle("igor")),
		mathjax.MathJax,
	),
)

type PageMarkdown struct {
	Root   ast.Node
	Header []byte
	Images []string
}

func ParseMarkdown(src []byte) PageMarkdown {
	var ret PageMarkdown
	ret.Root = markdown.Parser().Parse(text.NewReader(src))
	ret.Header = findHeader(ret.Root, src)
	ret.Images = getImages(ret.Root, src)
	return ret
}

func findHeader(root ast.Node, src []byte) []byte {
	var buf bytes.Buffer
	ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if n.Kind() == extensions.KindHeaderBlock {
			for i := 0; i < n.Lines().Len(); i++ {
				line := n.Lines().At(i)
				buf.Write(src[line.Start:line.Stop])
			}
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return buf.Bytes()
}

func getImages(root ast.Node, src []byte) []string {
	var imgs []string
	ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if n.Kind() == ast.KindImage {
			imgs = append(imgs, string(n.(*ast.Image).Destination))
		}
		return ast.WalkContinue, nil
	})
	return imgs
}

// Render renders the page in HTML format.
func RenderMarkdown(w io.Writer, source []byte, root ast.Node) error {
	return markdown.Renderer().Render(w, source, root)
}
