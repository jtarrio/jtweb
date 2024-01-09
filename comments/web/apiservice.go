package web

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"jacobo.tarrio.org/jtweb/comments"
	"jacobo.tarrio.org/jtweb/comments/service"
)

type apiService struct {
	service       service.CommentsService
	adminChecker  *AdminChecker
	handlers      map[handlerPath]http.HandlerFunc
	renderLimiter *rate.Limiter
}

type handlerPath struct {
	prefix string
	method string
	admin  bool
}

func userGet(path string) handlerPath {
	return handlerPath{prefix: path, method: http.MethodGet, admin: false}
}

func userPost(path string) handlerPath {
	return handlerPath{prefix: path, method: http.MethodPost, admin: false}
}

func adminPost(path string) handlerPath {
	return handlerPath{prefix: path, method: http.MethodPost, admin: true}
}

func Serve(service service.CommentsService, adminChecker *AdminChecker) http.Handler {
	out := &apiService{
		service:       service,
		adminChecker:  adminChecker,
		renderLimiter: rate.NewLimiter(1, 10),
	}
	out.handlers = map[handlerPath]http.HandlerFunc{
		userGet("/list/"):   out.list,
		userPost("/add"):    out.add,
		userPost("/render"): out.render,

		adminPost("/find"):       out.find,
		adminPost("/setVisible"): out.setVisible,
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

func (s *apiService) render(rw http.ResponseWriter, req *http.Request) {
	if limitRate(s.renderLimiter, rw) != nil {
		return
	}
	var inputData struct{ Text comments.Markdown }
	if input(req, &inputData, rw) != nil {
		return
	}
	outputData, err := s.service.Render(comments.Markdown(inputData.Text))
	output(struct{ Text comments.Html }{Text: outputData}, err, rw)
}

func (s *apiService) find(rw http.ResponseWriter, req *http.Request) {
	var params struct {
		Filter service.Filter
		Sort   service.Sort
		Limit  int
		Start  int
	}
	if input(req, &params, rw) != nil {
		return
	}
	result, err := s.service.Find(params.Filter, params.Sort, params.Limit, params.Start)
	output(result, err, rw)
}

func (s *apiService) setVisible(rw http.ResponseWriter, req *http.Request) {
	var params struct {
		Ids     map[service.PostId][]*service.CommentId
		Visible bool
	}
	if input(req, &params, rw) != nil {
		return
	}
	err := s.service.BulkSetVisible(params.Ids, params.Visible)
	output("Success", err, rw)
}

func (s *apiService) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for path, handler := range s.handlers {
		if req.Method != path.method {
			continue
		}
		newReq, found := stripPathPrefix(req, path.prefix)
		if !found {
			continue
		}
		if path.admin && !s.isAdmin(req) {
			continue
		}
		handler(rw, newReq)
		return
	}
	badRequest("invalid URL", rw)
}

func (s *apiService) isAdmin(req *http.Request) bool {
	return s.adminChecker.HasAdmin(req)
}

func limitRate(limiter *rate.Limiter, rw http.ResponseWriter) error {
	allowed := float64(limiter.Limit() * 60)
	remaining := limiter.Tokens()
	rw.Header().Add("X-RateLimit-Limit", fmt.Sprintf("%.0f", math.Floor(allowed)))
	rw.Header().Add("X-RateLimit-Remaining", fmt.Sprintf("%.0f", math.Floor(remaining)))
	resv := limiter.Reserve()
	delay := resv.Delay()
	if delay.Nanoseconds() != 0 {
		rw.Header().Add("Retry-After", fmt.Sprint(int(math.Ceil(delay.Seconds()))))
		rw.WriteHeader(http.StatusTooManyRequests)
		resv.Cancel()
		return fmt.Errorf("rate limit exceeded, retry after %f seconds", delay.Seconds())
	}
	return nil
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
