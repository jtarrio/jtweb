package extensions

import (
	"bytes"
	"fmt"
	"jtweb/page"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"gopkg.in/yaml.v2"
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

// RawHeader contains the structure for the header contents.
type RawHeader struct {
	Title           string
	Language        string
	Summary         string
	PublishDate     string `yaml:"publish_date"`
	HidePublishDate bool   `yaml:"no_publish_date"`
	AuthorName      string `yaml:"author_name"`
	AuthorURI       string `yaml:"author_uri"`
	HideAuthor      bool   `yaml:"hide_author"`
	Tags            []string
	NoIndex         bool     `yaml:"no_index"`
	OldURI          []string `yaml:"old_uris"`
	TranslationOf   string   `yaml:"translation_of"`
	Draft           bool
}

// ParseContents parses the header and returns a HeaderData object.
func (n *HeaderBlock) ParseContents(src []byte) (page.HeaderData, error) {
	var out page.HeaderData
	var buf bytes.Buffer
	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		buf.Write(src[line.Start:line.Stop])
	}
	rawHeader := &RawHeader{}
	err := yaml.UnmarshalStrict(buf.Bytes(), rawHeader)
	if err != nil {
		return out, err
	}

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
	out.Draft = rawHeader.Draft
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
