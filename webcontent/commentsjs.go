package webcontent

import (
	_ "embed"
	"net/http"
	"strings"
	"time"
)

//go:embed generated/comments.js
var commentsJs string

func ServeCommentsJs() http.HandlerFunc {
	now := time.Now()
	return func(rw http.ResponseWriter, req *http.Request) {
		sr := strings.NewReader(commentsJs)
		http.ServeContent(rw, req, "comments.js", now, sr)
	}
}
