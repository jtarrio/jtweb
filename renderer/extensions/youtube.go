package extensions

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type youTubeExtension struct{}

// YouTubeExtension lets you parse !youtube() blocks
var YouTubeExtension = &youTubeExtension{}

func (h *youTubeExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(newYouTubeBlockParser(), 10),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(newYouTubeBlockRenderer(), 10),
	))
}

// KindYouTubeBlock is the definition of the YouTubeBlock kind.
var KindYouTubeBlock = ast.NewNodeKind("YouTubeBlock")

// YouTubeBlock contains the Video ID referred to in a !youtube() block.
type YouTubeBlock struct {
	ast.BaseBlock
	VideoID string
}

func newYouTubeBlock(videoID string) *YouTubeBlock {
	return &YouTubeBlock{
		BaseBlock: ast.BaseBlock{},
		VideoID:   videoID,
	}
}

// Dump shows the contents of the block.
func (n *YouTubeBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"VideoID": n.VideoID}, nil)
}

// IsRaw tells whether the block contents should be rendered as-is.
func (n *YouTubeBlock) IsRaw() bool {
	return false
}

// Kind returns the kind definition for the block.
func (n *YouTubeBlock) Kind() ast.NodeKind {
	return KindYouTubeBlock
}

type youTubeBlockParser struct{}

func newYouTubeBlockParser() *youTubeBlockParser {
	return &youTubeBlockParser{}
}

func (p *youTubeBlockParser) Trigger() []byte {
	return []byte{'!'}
}

var lowerCaseTag = []byte{'y', 'o', 'u', 't', 'u', 'b', 'e'}
var upperCaseTag = []byte{'Y', 'O', 'U', 'T', 'U', 'B', 'E'}

func parseYouTubeLine(line []byte) ([]byte, bool) {
	line = util.TrimLeftSpace(util.TrimRightSpace(line))
	linelen := len(line)
	taglen := len(lowerCaseTag)
	minlen := taglen + 3
	if linelen < minlen || line[0] != '!' || line[taglen+1] != '(' || line[linelen-1] != ')' {
		return nil, false
	}
	for i := 0; i < taglen; i++ {
		if line[i+1] != lowerCaseTag[i] && line[i+1] != upperCaseTag[i] {
			return nil, false
		}
	}
	return bytes.TrimSpace(line[taglen+2 : linelen-1]), true
}

func getVideoID(uri string) string {
	question := strings.IndexByte(uri, '?')
	if question >= 0 {
		uri = uri[:question]
	}
	lastSlash := strings.LastIndexByte(uri, '/')
	if lastSlash >= 0 {
		uri = uri[lastSlash+1:]
	}
	return uri
}

func (p *youTubeBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	uri, ok := parseYouTubeLine(line)
	if !ok {
		return nil, parser.NoChildren
	}
	videoID := getVideoID(string(uri))
	reader.Advance(segment.Len() - 1)
	return newYouTubeBlock(videoID), parser.NoChildren
}

func (p *youTubeBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

func (p *youTubeBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (p *youTubeBlockParser) CanInterruptParagraph() bool {
	return false
}

func (p *youTubeBlockParser) CanAcceptIndentedLine() bool {
	return false
}

type youTubeBlockRenderer struct{}

func newYouTubeBlockRenderer() *youTubeBlockRenderer {
	return &youTubeBlockRenderer{}
}

func (r *youTubeBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindYouTubeBlock, r.renderYouTube)
}

func (r *youTubeBlockRenderer) renderYouTube(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*YouTubeBlock)
	width := 560
	height := 9 * width / 16
	_, _ = w.WriteString(fmt.Sprintf("<iframe class=\"youtube\" width=\"%d\" height=\"%d\" src=\"https://www.youtube.com/embed/%s\"", width, height, n.VideoID))
	_, _ = w.WriteString(" frameborder=\"0\" allow=\"accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture\" allowfullscreen></iframe>")
	return ast.WalkSkipChildren, nil
}
