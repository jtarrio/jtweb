package page

import (
	"time"

	"github.com/yuin/goldmark/ast"
)

// Page contains all the information about a parsed page.
type Page struct {
	Name   string
	Source []byte
	Root   ast.Node
	Header HeaderData
}

// HeaderData contains the information held in the page's header.
type HeaderData struct {
	Title           string
	Language        string
	Summary         string
	PublishDate     time.Time
	HidePublishDate bool
	AuthorName      string
	AuthorURI       string
	HideAuthor      bool
	Tags            []string
	NoIndex         bool
	OldURI          []string
	TranslationOf   string
}
