package webcontent

import (
	_ "embed"
	"net/http"
	"strings"
	"time"
)

//go:embed generated/comments.js
var commentsJs string

//go:embed html/admin.html
var adminHtml string

//go:embed generated/admin.js
var adminJs string

func ServeCommentsJs() http.HandlerFunc {
	return serveContent("comments.js", commentsJs)
}

func ServeAdminHtml() http.HandlerFunc {
	return serveContent("admin.html", adminHtml)
}

func ServeAdminJs() http.HandlerFunc {
	return serveContent("admin.js", adminJs)
}

func serveContent(name string, content string) http.HandlerFunc {
	now := time.Now()
	return func(rw http.ResponseWriter, req *http.Request) {
		sr := strings.NewReader(content)
		http.ServeContent(rw, req, name, now, sr)
	}
}
