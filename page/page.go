package page

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/yuin/goldmark/ast"
	"gopkg.in/yaml.v3"
	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/renderer"
)

// Name is a custom type for a page's identifier.
type Name string

// Page contains all the information about a parsed page.
type Page struct {
	Name   Name
	Source []byte
	Root   ast.Node
	Header HeaderData
}

// HeaderData contains the information held in the page's header.
type HeaderData struct {
	Title           string
	Language        languages.Language
	Summary         string
	Episode         string
	PublishDate     time.Time
	HidePublishDate bool
	AuthorName      string
	AuthorURI       string
	HideAuthor      bool
	CoverImage      string
	Tags            []string
	NoIndex         bool
	OldURI          []string
	TranslationOf   Name
	Comments        *CommentConfig
	Draft           bool
}

type CommentConfig struct {
	Enabled  bool
	Writable bool
}

// ParseCommentConfig parses the string given to the "comments" field
func ParseCommentConfig(enabled string) (*CommentConfig, error) {
	value := strings.ToLower(strings.TrimSpace(enabled))
	if value == "" {
		return nil, nil
	}
	if value == "false" || value == "off" || value == "no" || value == "disabled" {
		return &CommentConfig{Enabled: false, Writable: false}, nil
	}
	if value == "true" || value == "on" || value == "yes" || value == "enabled" || value == "open" {
		return &CommentConfig{Enabled: true, Writable: true}, nil
	}
	if value == "closed" || value == "readonly" {
		return &CommentConfig{Enabled: true, Writable: false}, nil
	}
	return nil, fmt.Errorf("unknown comment configuration: %s", enabled)
}

// Parse reads a page in Markdown format and parses it.
func Parse(name string, r io.Reader) (*Page, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	src := buf.Bytes()
	md := renderer.ParseMarkdown(src)
	header, err := parseHeader(md.Header, md.Images)
	if err != nil {
		return nil, err
	}
	page := &Page{Name: Name(name), Source: src, Root: md.Root, Header: header}
	return page, nil
}

// Render renders the page in HTML format.
func (p *Page) Render(w io.Writer) error {
	return renderer.RenderMarkdown(w, p.Source, p.Root)
}

// parseHeader parses the raw header and returns a HeaderData object.
func parseHeader(hdr []byte, imgs []string) (HeaderData, error) {
	var out HeaderData

	// RawHeader contains the structure for the header contents.
	type headerYaml struct {
		Title           string
		Language        string
		Summary         string
		Episode         string
		PublishDate     string  `yaml:"publish_date"`
		HidePublishDate bool    `yaml:"no_publish_date"`
		AuthorName      string  `yaml:"author_name"`
		AuthorURI       string  `yaml:"author_uri"`
		HideAuthor      bool    `yaml:"hide_author"`
		CoverImage      *string `yaml:"cover_image"`
		Tags            []string
		NoIndex         bool     `yaml:"no_index"`
		OldURI          []string `yaml:"old_uris"`
		TranslationOf   string   `yaml:"translation_of"`
		Comments        string
		Draft           bool
	}

	rawHeader := headerYaml{}
	decoder := yaml.NewDecoder(bytes.NewReader(hdr))
	decoder.KnownFields(true)
	err := decoder.Decode(&rawHeader)
	if err != nil {
		return out, err
	}

	if rawHeader.Title == "" {
		return out, fmt.Errorf("missing title")
	}

	out.Title = rawHeader.Title
	if rawHeader.Language == "" {
		out.Language = languages.LanguageEn
	} else {
		l, err := languages.FindByCode(rawHeader.Language)
		if err != nil {
			return out, err
		}
		out.Language = l
	}
	out.Summary = rawHeader.Summary
	if rawHeader.PublishDate != "" {
		d, err := parseDate(rawHeader.PublishDate)
		if err != nil {
			return HeaderData{}, err
		}
		out.PublishDate = d
	}
	out.Episode = rawHeader.Episode
	out.HidePublishDate = rawHeader.HidePublishDate
	out.AuthorName = rawHeader.AuthorName
	out.AuthorURI = rawHeader.AuthorURI
	out.HideAuthor = rawHeader.HideAuthor
	if rawHeader.CoverImage == nil {
		if len(imgs) > 0 {
			out.CoverImage = imgs[0]
		}
	} else {
		out.CoverImage = *rawHeader.CoverImage
	}
	out.Tags = rawHeader.Tags
	out.NoIndex = rawHeader.NoIndex
	out.OldURI = rawHeader.OldURI
	out.TranslationOf = Name(rawHeader.TranslationOf)
	cmtCfg, err := ParseCommentConfig(rawHeader.Comments)
	if err != nil {
		return HeaderData{}, err
	}
	out.Comments = cmtCfg
	out.Draft = rawHeader.Draft
	return out, nil
}

var dateFormats = []string{"2006-01-02", "01/02/2006", "02-01-2006", "20060102"}
var timeFormats = []string{" 3:04:05pm", " 3:04pm", " 3:04:05PM", " 3:04PM", " 15:04:05", " 15:04", "150405", "1504", "T15:04:05", "T15:04"}
var tzFormats = []string{" -07:00 MST", " -0700 MST", " -07:00", " -0700", "Z07:00", "Z0700"}

func parseDate(str string) (time.Time, error) {
	for _, df := range dateFormats {
		d, err := time.Parse(df, str)
		if err == nil {
			return d, nil
		}
		for _, tf := range timeFormats {
			dt, err := time.Parse(df+tf, str)
			if err == nil {
				return dt, nil
			}
			for _, zf := range tzFormats {
				dtz, err := time.Parse(df+tf+zf, str)
				if err == nil {
					return dtz, nil
				}
			}
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format for %s", str)
}
