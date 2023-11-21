package extensions

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type headerExtension struct{}

// HeaderExtension lets you parse HEADER blocks.
var HeaderExtension = &headerExtension{}

func (h *headerExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(newHeaderBlockParser(), 0),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(newHeaderBlockRenderer(), 0),
	))
}

// KindHeaderBlock designates blocks containing headers.
var KindHeaderBlock = ast.NewNodeKind("HeaderBlock")

// HeaderBlock is a block containing header data.
type HeaderBlock struct {
	ast.BaseBlock
}

func newHeaderBlock() *HeaderBlock {
	return &HeaderBlock{BaseBlock: ast.BaseBlock{}}
}

// Dump outputs the node.
func (n *HeaderBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{}, nil)
}

// IsRaw tells whether the node contents should be written in raw mode.
func (n *HeaderBlock) IsRaw() bool {
	return true
}

// Kind returns the node's kind.
func (n *HeaderBlock) Kind() ast.NodeKind {
	return KindHeaderBlock
}

type headerBlockParser struct{}

func newHeaderBlockParser() *headerBlockParser {
	return &headerBlockParser{}
}

func (p *headerBlockParser) Trigger() []byte {
	return []byte{'<'}
}

func (p *headerBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	ofs := pc.BlockOffset()
	if ofs < 0 {
		return nil, parser.NoChildren
	}
	if !bytes.HasPrefix(line[ofs:], []byte("<!--HEADER")) {
		return nil, parser.NoChildren
	}
	reader.Advance(segment.Len() - 1)
	return newHeaderBlock(), parser.NoChildren
}

func (p *headerBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	spaces := util.TrimLeftSpaceLength(line)
	if bytes.HasPrefix(line[spaces:], []byte("-->")) {
		reader.Advance(spaces + 3)
		return parser.Close
	}
	node.Lines().Append(segment)
	reader.Advance(segment.Len() - 1)
	return parser.Continue | parser.NoChildren
}

func (p *headerBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (p *headerBlockParser) CanInterruptParagraph() bool {
	return false
}

func (p *headerBlockParser) CanAcceptIndentedLine() bool {
	return false
}

type headerBlockRenderer struct{}

func newHeaderBlockRenderer() *headerBlockRenderer {
	return &headerBlockRenderer{}
}

func (r *headerBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindHeaderBlock, r.renderHeader)
}

func (r *headerBlockRenderer) renderHeader(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
