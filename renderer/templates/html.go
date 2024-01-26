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

func HtmlToText(i io.Reader, o io.Writer, linksTitle string, pictureTitle string) error {
	var t text
	var links []string
	doc, err := html.Parse(i)
	if err != nil {
		return err
	}
	type listItem struct {
		number *int
	}
	type processParams struct {
		inPara, inPre bool
		lists         []listItem
	}
	var processNode func(*html.Node, processParams)
	processNode = func(n *html.Node, params processParams) {
		if n.Type == html.ElementNode {
			if n.Data == "p" {
				t.startParagraph()
				params.inPara = true
			} else if n.Data == "pre" {
				t.startParagraph()
				params.inPre = true
			} else if n.Data == "img" {
				caption := ""
				for _, a := range n.Attr {
					if a.Key == "alt" || (a.Key == "src" && caption == "") {
						caption = a.Val
					}
				}
				if caption != "" {
					if !params.inPara {
						t.startParagraph()
					}
					t.appendWrapping("(" + pictureTitle + ": " + caption + ")")
				}
			} else if n.Data == "sup" {
				t.append("^(")
			} else if n.Data == "hr" {
				t.startParagraph()
				t.append("***")
				t.startParagraph()
			} else if n.Data == "ol" {
				t.startParagraph()
				number := 1
				params.lists = append(params.lists, listItem{number: &number})
			} else if n.Data == "ul" {
				t.startParagraph()
				params.lists = append(params.lists, listItem{number: nil})
			} else if n.Data == "li" {
				t.startParagraph()
				li := params.lists[len(params.lists)-1]
				if li.number == nil {
					t.prefix("  *  ")
				} else {
					t.prefix(fmt.Sprintf(" %2d. ", *li.number))
					(*li.number)++
				}
				t.addIndent()
			}
		} else if n.Type == html.TextNode {
			if params.inPara {
				t.appendWrapping(n.Data)
			} else if params.inPre {
				t.append(n.Data)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processNode(c, params)
		}
		if n.Type == html.ElementNode {
			if n.Data == "a" {
				for _, a := range n.Attr {
					if a.Key == "href" {
						if !isFootnoteLink(a.Val) {
							links = append(links, a.Val)
							t.appendWrapping(string(zeroWidthSpace) + "[" + fmt.Sprint(len(links)) + "]")
						}
					}
				}
			} else if n.Data == "sup" {
				t.append(")")
			} else if n.Data == "ol" || n.Data == "ul" {
				params.lists = params.lists[:len(params.lists)-1]
			} else if n.Data == "li" {
				t.removeIndent()
			}
		}
	}
	processNode(doc, processParams{
		inPara: false,
		inPre:  false,
		lists:  []listItem{},
	})

	if len(links) > 0 {
		t.addLine()
		t.addLine()
		t.appendWrapping(linksTitle + ":")
		for i, l := range links {
			t.addLine()
			t.append(fmt.Sprintf("  [%d] %s", i+1, l))
		}
	}

	_, err = io.WriteString(o, strings.Join(t.lines, "\n"))
	return err
}
