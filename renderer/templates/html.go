package templates

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func makeAbsolute(base *url.URL, path *url.URL, uri string) string {
	if strings.HasPrefix(uri, "{$") {
		return uri
	}
	lastPath, err := url.Parse(uri)
	if err != nil {
		return uri
	}
	if lastPath.Scheme == "" &&
		lastPath.Host == "" &&
		lastPath.User == nil &&
		lastPath.Path != "" &&
		lastPath.Path[0] == '/' {
		lastPath.Path = lastPath.Path[1:]
		return base.ResolveReference(lastPath).String()
	}
	return base.ResolveReference(path).ResolveReference(lastPath).String()
}

// MakeUrisAbsolute reads an HTML document from `i`, replaces all relative URIs
// with absolute, and writes the result to `o`.
//
// Relative URIs are referenced to the concatenation of `base` and `path`;
// site-relative URIs are referenced to `base` as if they were relative;
// absolute URIs are passed as-is.
func MakeUrisAbsolute(i io.Reader, o io.Writer, base string, path string) error {
	baseUri, err := url.Parse(base)
	if err != nil {
		return err
	}
	pathUri, err := url.Parse(path)
	if err != nil {
		return err
	}
	doc, err := html.Parse(i)
	if err != nil {
		return err
	}
	var processNode func(*html.Node)
	processNode = func(n *html.Node) {
		if n.Type == html.ElementNode {
			wantedAttr := ""
			if n.Data == "a" {
				wantedAttr = "href"
			} else if n.Data == "img" {
				wantedAttr = "src"
			}
			if wantedAttr != "" {
				for i, a := range n.Attr {
					if a.Key == wantedAttr {
						n.Attr[i].Val = makeAbsolute(baseUri, pathUri, a.Val)
						break
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processNode(c)
		}
	}
	processNode(doc)
	return html.Render(o, doc)
}

const textWrapColumns = 79
const zeroWidthSpace rune = '\u200b'

type text struct {
	lines     []string
	line      *string
	lineLen   int
	separator rune
}

func (t *text) addLine() {
	t.lines = append(t.lines, "")
	t.lineLen = 0
	t.line = &t.lines[len(t.lines)-1]
	t.separator = 0
}

func (t *text) startParagraph() {
	if t.line != nil && *t.line != "" {
		t.addLine()
		t.addLine()
	}
}

func (t *text) append(n string) {
	if t.line == nil {
		t.addLine()
	}
	*t.line += n
	t.lineLen += len([]rune(n))
	t.separator = 0
}

func cutSeparators(s string) (first string, rest string, sep rune) {
	for i, c := range s {
		if c == ' ' || c == '\n' || c == zeroWidthSpace {
			return s[:i], s[i+len(string(c)):], c
		}
	}
	return s, "", 0
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
				*t.line = first
				t.lineLen = firstLen
			} else if t.lineLen == 0 || sep == zeroWidthSpace || t.separator == 0 || t.separator == zeroWidthSpace {
				*t.line += first
				t.lineLen += firstLen
			} else {
				*t.line += " " + first
				t.lineLen += firstLen + 1
			}
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
	var processNode func(*html.Node, bool, bool)
	processNode = func(n *html.Node, inPara, inPre bool) {
		if n.Type == html.ElementNode {
			if n.Data == "p" {
				t.startParagraph()
				inPara = true
			} else if n.Data == "pre" {
				t.startParagraph()
				inPre = true
			} else if n.Data == "img" {
				caption := ""
				for _, a := range n.Attr {
					if a.Key == "alt" || (a.Key == "src" && caption == "") {
						caption = a.Val
					}
				}
				if caption != "" {
					if !inPara {
						t.startParagraph()
					}
					t.appendWrapping("(" + pictureTitle + ": " + caption + ")")
				}
			}
		} else if n.Type == html.TextNode {
			if inPara {
				t.appendWrapping(n.Data)
			} else if inPre {
				t.append(n.Data)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processNode(c, inPara, inPre)
		}
		if n.Type == html.ElementNode {
			if n.Data == "a" {
				for _, a := range n.Attr {
					if a.Key == "href" {
						links = append(links, a.Val)
						t.appendWrapping(string(zeroWidthSpace) + "[" + fmt.Sprint(len(links)) + "]")
					}
				}
			}
		}
	}
	processNode(doc, false, false)

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
