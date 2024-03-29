package renderer

import (
	"bytes"
	"io"
	"net/url"
	"regexp"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
)

func RewriteUrl(siteUrl, pageUrl, targetUrl string) (string, error) {
	base, err := url.Parse(siteUrl)
	if err != nil {
		return "", err
	}
	page, err := url.Parse(pageUrl)
	if err != nil {
		return "", err
	}
	target, err := url.Parse(targetUrl)
	if err != nil {
		return "", err
	}
	return rewriteUrl(base, page, target).String(), nil
}

func rewriteUrl(base, page, target *url.URL) *url.URL {
	if target.Scheme == "" && target.Host == "" && target.User == nil && target.Path == "" && target.RawQuery == "" && target.Fragment != "" {
		return target
	}
	if target.Scheme == "" && target.Host == "" && target.User == nil && target.Path != "" && target.Path[0] == '/' {
		target.Path = target.Path[1:]
		return base.ResolveReference(target)
	}
	return base.ResolveReference(page).ResolveReference(target)
}

func makeUrlRewriter(siteUrl, pageUrl string) (func(u *url.URL), error) {
	base, err := url.Parse(siteUrl)
	if err != nil {
		return nil, err
	}
	page, err := url.Parse(pageUrl)
	if err != nil {
		return nil, err
	}
	rewriter := func(u *url.URL) {
		*u = *rewriteUrl(base, page, u)
	}
	return rewriter, nil
}

func rewriteUrls(w io.Writer, r io.Reader, rewriter func(u *url.URL)) error {
	rewrite := func(s string) string {
		u, err := url.Parse(s)
		if err != nil {
			return ""
		}
		rewriter(u)
		return u.String()
	}

	tokenizer := html.NewTokenizer(r)
	for {
		if tokenizer.Next() == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				return nil
			}
			return err
		}
		token := tokenizer.Token()
		if token.Type == html.StartTagToken || token.Type == html.SelfClosingTagToken {
			switch token.Data {
			case "a", "area", "base", "link":
				for i, attr := range token.Attr {
					if attr.Key == "href" {
						attr.Val = rewrite(attr.Val)
						token.Attr[i] = attr
					}
				}
			case "audio", "embed", "iframe", "img", "script", "source", "track", "video":
				for i, attr := range token.Attr {
					if attr.Key == "src" {
						attr.Val = rewrite(attr.Val)
						token.Attr[i] = attr
					}
				}
			}
		}
		w.Write([]byte(token.String()))
	}
}

func SanitizePost(w io.Writer, r io.Reader, siteUrl, pageUrl string) error {
	rewriter, err := makeUrlRewriter(siteUrl, pageUrl)
	if err != nil {
		return err
	}
	buf := bytes.Buffer{}
	err = rewriteUrls(&buf, r, rewriter)
	if err != nil {
		return err
	}

	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").Matching(bluemonday.SpaceSeparatedTokens).OnElements("div", "span")
	p.AllowAttrs("title").OnElements("a", "img")
	p.AllowAttrs("alt").OnElements("img")
	p.AllowAttrs("style").Globally()
	p.AllowAttrs("src", "class", "width", "height", "sandbox").OnElements("iframe")
	p.AllowAttrs("frameborder").Matching(regexp.MustCompile(`^0$`)).OnElements("iframe")
	p.RequireNoFollowOnLinks(false)
	return p.SanitizeReaderToWriter(&buf, w)
}

func NormalizeOutput(w io.Writer, r io.Reader) error {
	doc, err := html.Parse(r)
	if err != nil {
		return err
	}
	return html.Render(w, doc)
}
