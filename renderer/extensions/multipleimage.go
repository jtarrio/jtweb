package extensions

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type multipleImageExtension struct{}

// MultipleImageExtension lets us render multiple images together.
var MultipleImageExtension = &multipleImageExtension{}

func (e *multipleImageExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(newMultipleImageASTTransformer(), 20)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(newMultipleImageRenderer(), 20)))
}

// KindMultipleImage designates a multiple image's node.
var KindMultipleImage = ast.NewNodeKind("MultipleImage")

// MultipleImage is a node containing multiple images.
type MultipleImage struct {
	ast.BaseInline
}

func newMultipleImage() *MultipleImage {
	return &MultipleImage{BaseInline: ast.BaseInline{}}
}

// Dump displays the current node.
func (n *MultipleImage) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// IsRaw tells whether the node's content is raw.
func (n *MultipleImage) IsRaw() bool {
	return false
}

// Kind returns the node's kind.
func (n *MultipleImage) Kind() ast.NodeKind {
	return KindMultipleImage
}

type multipleImageASTTransformer struct{}

func newMultipleImageASTTransformer() *multipleImageASTTransformer {
	return &multipleImageASTTransformer{}
}

func (t *multipleImageASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	imgs := make([]ast.Node, 0)
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindImage {
			prev := previousNonBlankSibling(n)
			next := nextNonBlankSibling(n)
			if (prev == nil || !isSingleImage(prev)) && next != nil && isSingleImage(next) {
				imgs = append(imgs, n)
			}
		}
		return ast.WalkContinue, nil
	})
	for _, img := range imgs {
		siblings := getNextImgs(img)
		nonBlank := getNonBlankNodes(siblings)
		if len(nonBlank) < 2 {
			continue
		}
		parent := img.Parent()
		multi := newMultipleImage()
		parent.InsertBefore(parent, img, multi)
		for _, sib := range siblings {
			parent.RemoveChild(parent, sib)
		}
		for _, sib := range nonBlank {
			multi.AppendChild(multi, sib)
		}
	}
}

func isSingleImage(n ast.Node) bool {
	if n.Kind() == ast.KindImage {
		return true
	}
	if n.Kind() == ast.KindLink {
		child := n.FirstChild()
		return child != nil && child.NextSibling() == nil && isSingleImage(child)
	}
	return false
}

func getNextImgs(n ast.Node) []ast.Node {
	sibs := []ast.Node{n}
	for sib := nextNonBlankSibling(n); sib != nil && isSingleImage(sib); sib = nextNonBlankSibling(sib) {
		sibs = append(sibs, sib)
	}
	return sibs
}

type multipleImageRenderer struct{}

func newMultipleImageRenderer() *multipleImageRenderer {
	return &multipleImageRenderer{}
}

func (r *multipleImageRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMultipleImage, r.renderMultipleImage)
}

func (r *multipleImageRenderer) renderMultipleImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<span class=\"multipleImgs\">")
	} else {
		w.WriteString("</span>")
	}
	return ast.WalkContinue, nil
}
