package web

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"jacobo.tarrio.org/jtweb/comments"
	"jacobo.tarrio.org/jtweb/comments/service"
)

type apiService struct {
	service  service.CommentsService
	handlers map[handlerPath]http.HandlerFunc
}

type handlerPath struct {
	prefix string
	method string
}

func Serve(service service.CommentsService) http.Handler {
	out := &apiService{service: service}
	out.handlers = map[handlerPath]http.HandlerFunc{
		{prefix: "/list/", method: http.MethodGet}: out.list,
		{prefix: "/add", method: http.MethodPost}:  out.add,
	}
	return out
}

func (s *apiService) list(rw http.ResponseWriter, req *http.Request) {
	postId := comments.PostId(strings.TrimPrefix(req.URL.Path, "/"))
	list, err := s.service.List(postId, false)
	output(list, err, rw)
}

func (s *apiService) add(rw http.ResponseWriter, req *http.Request) {
	var newComment service.NewComment
	if input(req, &newComment, rw) != nil {
		return
	}
	newComment.When = time.Now()
	comment, err := s.service.Add(&newComment)
	output(comment, err, rw)
}

func (s *apiService) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for path, handler := range s.handlers {
		if req.Method != path.method {
			continue
		}
		newReq, found := stripPathPrefix(req, path.prefix)
		if found {
			handler(rw, newReq)
			return
		}
	}
	badRequest("invalid URL", rw)
}

func output(what any, err error, rw http.ResponseWriter) {
	if err == nil {
		var output []byte
		output, err = json.Marshal(what)
		if err == nil {
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

func stripPathPrefix(req *http.Request, prefix string) (newReq *http.Request, found bool) {
	after, ok := strings.CutPrefix(req.URL.Path, prefix)
	if !ok {
		return req, false
	}
	newReq = new(http.Request)
	*newReq = *req
	newReq.URL = new(url.URL)
	newReq.URL.Path = after
	newReq.URL.RawPath = strings.TrimPrefix(req.URL.RawPath, prefix)
	return newReq, true
}
