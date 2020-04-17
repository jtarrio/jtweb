package extensions

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type imageCaptionExtension struct{}

// ImageCaptionExtension adds a caption to images, formed by whatever text follows the image.
var ImageCaptionExtension = &imageCaptionExtension{}

func (e *imageCaptionExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(newImageCaptionASTTransformer(), 21)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(newImageCaptionRenderer(), 21)))
}

// KindImageCaption designates an image caption's node.
var KindImageCaption = ast.NewNodeKind("ImageCaption")

// ImageCaption is a node containing an image's caption.
type ImageCaption struct {
	ast.BaseInline
}

func newImageCaption() *ImageCaption {
	return &ImageCaption{BaseInline: ast.BaseInline{}}
}

// Dump displays the current node.
func (n *ImageCaption) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// IsRaw tells whether the node's content is raw.
func (n *ImageCaption) IsRaw() bool {
	return false
}

// Kind returns the node's kind.
func (n *ImageCaption) Kind() ast.NodeKind {
	return KindImageCaption
}

type imageCaptionASTTransformer struct{}

func newImageCaptionASTTransformer() *imageCaptionASTTransformer {
	return &imageCaptionASTTransformer{}
}

func (t *imageCaptionASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	imgs := make([]ast.Node, 0)
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && isImage(n) && nextNonBlankSibling(n) != nil {
			imgs = append(imgs, n)
		}
		return ast.WalkContinue, nil
	})
	for _, img := range imgs {
		siblings := getCaptionSiblings(img)
		nonBlank := getNonBlankNodes(siblings)
		if len(nonBlank) == 0 {
			continue
		}
		parent := img.Parent()
		caption := newImageCaption()
		parent.InsertAfter(parent, img, caption)
		for _, sib := range siblings {
			parent.RemoveChild(parent, sib)
		}
		for _, sib := range nonBlank {
			caption.AppendChild(caption, sib)
		}
	}
}

func isImage(n ast.Node) bool {
	return n.Kind() == ast.KindImage || n.Kind() == KindMultipleImage
}

func getCaptionSiblings(n ast.Node) []ast.Node {
	sibs := make([]ast.Node, 0)
	for sib := n.NextSibling(); sib != nil && !isImage(sib); sib = sib.NextSibling() {
		sibs = append(sibs, sib)
	}
	return sibs
}

type imageCaptionRenderer struct{}

func newImageCaptionRenderer() *imageCaptionRenderer {
	return &imageCaptionRenderer{}
}

func (r *imageCaptionRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindImageCaption, r.renderImageCaption)
}

func (r *imageCaptionRenderer) renderImageCaption(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("<span class=\"imageCaption\">")
	} else {
		w.WriteString("</span>")
	}
	return ast.WalkContinue, nil
}
