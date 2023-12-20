package web

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"jacobo.tarrio.org/jtweb/comments"
	"jacobo.tarrio.org/jtweb/comments/service"
)

type webService struct {
	service service.CommentsService
}

func Serve(service service.CommentsService) http.Handler {
	return &webService{service: service}
}

func output(what any, err error, rw http.ResponseWriter) {
	if err == nil {
		var output []byte
		output, err = json.Marshal(what)
		if err != nil {
			rw.WriteHeader(http.StatusOK)
			rw.Header().Add("Content-Type", "application/json")
			rw.Write(output)
			return
		}
	}
	rw.WriteHeader(http.StatusInternalServerError)
	rw.Header().Add("Content-Type", "text/plain")
	rw.Write([]byte(err.Error()))
}

func input(req *http.Request, v any, rw http.ResponseWriter) error {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err == nil {
		err = json.Unmarshal(body, v)
	}
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Header().Add("Content-Type", "text/plain")
		rw.Write([]byte(err.Error()))
		return err
	}
	return nil
}

func badRequest(text string, rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusBadRequest)
	rw.Header().Add("Content-Type", "text/plain")
	rw.Write([]byte(text))
}

func hasMethod(method string, req *http.Request, rw http.ResponseWriter) bool {
	if req.Method != method {
		badRequest("invalid method", rw)
		return false
	}
	return true
}

func (s *webService) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	{
		after, ok := strings.CutPrefix(req.URL.Path, "/get/")
		if ok {
			if !hasMethod(http.MethodGet, req, rw) {
				return
			}
			postId := comments.PostId(after)
			list, err := s.service.Get(postId)
			output(list, err, rw)
			return
		}
	}
	{
		if req.URL.Path == "/add" {
			if !hasMethod(http.MethodPost, req, rw) {
				return
			}
			var newComment service.NewComment
			if input(req, &newComment, rw) != nil {
				return
			}
			comment, err := s.service.Add(&newComment)
			output(comment, err, rw)
			return
		}
	}
	badRequest("invalid URL", rw)
}
