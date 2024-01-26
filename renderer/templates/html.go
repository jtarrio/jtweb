package templates

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

const textWrapColumns = 79
const zeroWidthSpace rune = '\u200b'

type text struct {
	lines     []string
	line      *string
	lineLen   int
	indent    int
	inPara    bool
	separator rune
}

func (t *text) addLine() {
	newLine := strings.Repeat("     ", t.indent)
	t.lines = append(t.lines, newLine)
	t.lineLen = len([]rune(newLine))
	t.line = &t.lines[len(t.lines)-1]
	t.inPara = false
	t.separator = 0
	if t.indent > 0 {
		t.separator = zeroWidthSpace
	}
}

func (t *text) startParagraph() {
	if t.inPara {
		t.addLine()
		t.addLine()
	}
}

func (t *text) append(n string) {
	lines := strings.Split(n, "\n")
	for i, line := range lines {
		if i > 0 || t.line == nil {
			t.addLine()
		}
		*t.line += line
		t.lineLen += len([]rune(line))
		t.separator = 0
		t.inPara = true
	}
}

func (t *text) prefix(n string) {
	t.append(n)
	t.inPara = false
	t.separator = zeroWidthSpace
}

func (t *text) addIndent() {
	t.indent++
}

func (t *text) removeIndent() {
	t.indent--
	if t.indent < 0 {
		t.indent = 0
	}
}

func cutSeparators(s string) (first string, rest string, sep rune) {
	for i, c := range s {
		if c == ' ' || c == '\n' || c == zeroWidthSpace {
			return s[:i], s[i+len(string(c)):], c
		}
	}
	return s, "", 0
}

func isFootnoteLink(url string) bool {
	_, after, found := strings.Cut(url, "#")
	if !found {
		return false
	}
	return strings.HasPrefix(after, "fn:") || strings.HasPrefix(after, "fnref:")
}

func (t *text) appendWrapping(n string) {
	for n != "" {
		first, rest, sep := cutSeparators(n)
		n = rest
		if first != "" {
			if t.line == nil {
				t.addLine()
			}
			firstLen := len([]rune(first))
			if t.lineLen > 0 && t.lineLen+1+firstLen > textWrapColumns {
				t.addLine()
				*t.line += first
				t.lineLen += firstLen
			} else if t.lineLen == 0 || sep == zeroWidthSpace || t.separator == 0 || t.separator == zeroWidthSpace {
				*t.line += first
				t.lineLen += firstLen
			} else {
				*t.line += " " + first
				t.lineLen += firstLen + 1
			}
			t.inPara = true
		}
		t.separator = sep
	}
}

type Titles struct {
	Links   string
	Picture string
	Notes   string
}

type htmlToTextParams struct {
	titles        Titles
	links         *[]string
	inPara, inPre bool
	toUpper       bool
	lists         []*uint
	inFootnotes   bool
}

type htmlTagProcessor func(n *html.Node, t *text, params *htmlToTextParams)

type htmlTag struct {
	preFn  htmlTagProcessor
	postFn htmlTagProcessor
}

func doNothing(*html.Node, *text, *htmlToTextParams) {}

func endAnchor(n *html.Node, t *text, params *htmlToTextParams) {
	for _, a := range n.Attr {
		if a.Key == "href" {
			if !isFootnoteLink(a.Val) {
				*params.links = append(*params.links, a.Val)
				t.appendWrapping(string(zeroWidthSpace) + "[" + fmt.Sprint(len(*params.links)) + "]")
			}
		}
	}
}

func startParagraph(n *html.Node, t *text, params *htmlToTextParams) {
	t.startParagraph()
	params.inPara = true
}

func startHeader(n *html.Node, t *text, params *htmlToTextParams) {
	t.startParagraph()
	params.inPara = true
	params.toUpper = n.Data == "h1"
	prefix := "   "
	if n.Data == "h2" {
		prefix = " * "
	} else if n.Data == "h3" {
		prefix = " = "
	} else if n.Data == "h4" {
		prefix = " - "
	}
	t.prefix(prefix)
}

func startPre(n *html.Node, t *text, params *htmlToTextParams) {
	t.startParagraph()
	params.inPre = true
}

func startImg(n *html.Node, t *text, params *htmlToTextParams) {
	caption := ""
	for _, a := range n.Attr {
		if a.Key == "alt" || (a.Key == "src" && caption == "") {
			caption = a.Val
		}
	}
	if caption == "" {
		return
	}
	if !params.inPara {
		t.startParagraph()
	}
	t.appendWrapping("(" + params.titles.Picture + ": " + caption + ")")
}

func startHr(n *html.Node, t *text, params *htmlToTextParams) {
	t.startParagraph()
	t.append("***")
	if params.inFootnotes {
		t.addIndent()
		t.startParagraph()
		t.appendWrapping(params.titles.Notes)
		t.removeIndent()
	}
	t.startParagraph()
}

func startOl(n *html.Node, t *text, params *htmlToTextParams) {
	t.startParagraph()
	var number uint = 1
	params.lists = append(params.lists, &number)
}

func startUl(n *html.Node, t *text, params *htmlToTextParams) {
	t.startParagraph()
	params.lists = append(params.lists, nil)
}

func endList(n *html.Node, t *text, params *htmlToTextParams) {
	params.lists = params.lists[:len(params.lists)-1]
}

func startLi(n *html.Node, t *text, params *htmlToTextParams) {
	t.startParagraph()
	li := params.lists[len(params.lists)-1]
	if li == nil {
		t.prefix("   * ")
	} else {
		t.prefix(fmt.Sprintf("%3d. ", *li))
		(*li)++
	}
	t.addIndent()
}

func endLi(n *html.Node, t *text, params *htmlToTextParams) {
	t.removeIndent()
}

func startDiv(n *html.Node, t *text, params *htmlToTextParams) {
	for _, a := range n.Attr {
		if a.Key == "class" && a.Val == "footnotes" {
			params.inFootnotes = true
		}
	}
}

func appendText(txt string) htmlTagProcessor {
	return func(n *html.Node, t *text, params *htmlToTextParams) {
		t.append(txt)
	}
}

var htmlTags = map[string]htmlTag{
	"a":   {doNothing, endAnchor},
	"p":   {startParagraph, doNothing},
	"h1":  {startHeader, doNothing},
	"h2":  {startHeader, doNothing},
	"h3":  {startHeader, doNothing},
	"h4":  {startHeader, doNothing},
	"h5":  {startHeader, doNothing},
	"h6":  {startHeader, doNothing},
	"pre": {startPre, doNothing},
	"img": {startImg, doNothing},
	"sup": {appendText("^("), appendText(")")},
	"hr":  {startHr, doNothing},
	"ol":  {startOl, endList},
	"ul":  {startUl, endList},
	"li":  {startLi, endLi},
	"div": {startDiv, doNothing},
}

func HtmlToText(i io.Reader, o io.Writer, titles Titles) error {
	var t text
	doc, err := html.Parse(i)
	if err != nil {
		return err
	}
	links := []string{}
	params := htmlToTextParams{
		titles: titles,
		links:  &links,
	}
	var processNode func(*html.Node, htmlToTextParams)
	processNode = func(n *html.Node, params htmlToTextParams) {
		if n.Type == html.TextNode {
			txt := n.Data
			if params.toUpper {
				txt = strings.ToUpper(txt)
			}
			if params.inPara {
				t.appendWrapping(txt)
			} else if params.inPre {
				t.append(txt)
			}
		} else if n.Type == html.ElementNode {
			tag, ok := htmlTags[n.Data]
			if !ok {
				tag = htmlTag{doNothing, doNothing}
			}
			tag.preFn(n, &t, &params)
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				processNode(c, params)
			}
			tag.postFn(n, &t, &params)
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				processNode(c, params)
			}
		}
	}
	processNode(doc, params)

	if len(links) > 0 {
		t.addLine()
		t.addLine()
		t.appendWrapping(titles.Links + ":")
		t.addLine()
		for i, l := range links {
			t.addLine()
			t.append(fmt.Sprintf("  [%d] %s", i+1, l))
		}
	}

	_, err = io.WriteString(o, strings.Join(t.lines, "\n"))
	return err
}
