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
	return base.ResolveReference(path).ResolveReference(lastPath).String()
}

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

type text struct {
	lines []string
	line  *string
}

func (t *text) isLastLineEmpty() bool {
	return t.line == nil || *t.line == ""
}

func (t *text) trimLastLine() {
	if !t.isLastLineEmpty() {
		*t.line = strings.TrimSpace(*t.line)
	}
}

func (t *text) addLine() {
	t.lines = append(t.lines, "")
	t.line = &(t.lines[len(t.lines)-1])
}

func (t *text) trimAndAddLine() {
	t.trimLastLine()
	t.addLine()
}

func wordWrap(input []string, columns int) []string {
	var t text
	for _, in := range input {
		t.trimAndAddLine()
		for {
			first, rest, found := strings.Cut(in, " ")
			if len(*t.line)+1+len(first) > columns {
				t.trimAndAddLine()
				*t.line = first
			} else {
				*t.line = *t.line + " " + first
			}
			in = rest
			if !found {
				break
			}
		}
	}
	t.trimLastLine()
	return t.lines
}

func HtmlToText(i io.Reader, o io.Writer, linksTitle string, pictureTitle string) error {
	var t text
	var links []string
	doc, err := html.Parse(i)
	if err != nil {
		return err
	}
	var processNode func(*html.Node, bool)
	processNode = func(n *html.Node, inPara bool) {
		if n.Type == html.ElementNode {
			if n.Data == "p" {
				if !t.isLastLineEmpty() {
					t.trimAndAddLine()
				}
				t.trimAndAddLine()
				inPara = true
			} else if n.Data == "img" {
				for _, a := range n.Attr {
					if a.Key == "alt" {
						if !t.isLastLineEmpty() {
							t.trimAndAddLine()
						}
						*t.line = "(" + pictureTitle + ": " + a.Val + ")"
						t.trimAndAddLine()
					}
				}
			}
		} else if n.Type == html.TextNode && inPara {
			*t.line = *t.line + n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processNode(c, inPara)
		}
		if n.Type == html.ElementNode {
			if n.Data == "a" {
				for _, a := range n.Attr {
					if a.Key == "href" {
						links = append(links, a.Val)
						*t.line = *t.line + "[" + fmt.Sprint(len(links)) + "]"
					}
				}
			}
		}
	}
	processNode(doc, false)
	t.trimLastLine()
	t.lines = wordWrap(t.lines, 79)

	if len(links) > 0 {
		t.trimAndAddLine()
		t.addLine()
		t.addLine()
		*t.line = linksTitle + ":"
		t.addLine()
		for i, l := range links {
			t.addLine()
			*t.line = fmt.Sprintf("  [%d] %s", i+1, l)
		}
	}

	_, err = io.WriteString(o, strings.Join(t.lines, "\n"))
	return err
}
