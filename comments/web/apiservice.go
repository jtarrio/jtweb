package web

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"jacobo.tarrio.org/jtweb/comments"
	"jacobo.tarrio.org/jtweb/comments/service"
)

type apiService struct {
	webService
	service       service.CommentsService
	renderLimiter *rate.Limiter
}

func Serve(service service.CommentsService, adminChecker *AdminChecker) http.Handler {
	out := &apiService{
		service:       service,
		renderLimiter: rate.NewLimiter(1, 10),
	}
	out.adminChecker = adminChecker
	out.handlers = map[handlerPath]http.HandlerFunc{
		userGet("/list/"):   out.list,
		userPost("/add"):    out.add,
		userPost("/render"): out.render,

		adminPost("/findComments"):          out.findComments,
		adminPost("/findPosts"):             out.findPosts,
		adminPost("/bulkSetVisible"):        out.bulkSetVisible,
		adminPost("/bulkUpdatePostConfigs"): out.bulkUpdatePostConfigs,
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

func (s *apiService) findComments(rw http.ResponseWriter, req *http.Request) {
	var params struct {
		Filter service.CommentFilter
		Sort   service.Sort
		Limit  int
		Start  int
	}
	if input(req, &params, rw) != nil {
		return
	}
	result, err := s.service.FindComments(params.Filter, params.Sort, params.Limit, params.Start)
	output(result, err, rw)
}

func (s *apiService) findPosts(rw http.ResponseWriter, req *http.Request) {
	var params struct {
		Filter service.PostFilter
		Sort   service.Sort
		Limit  int
		Start  int
	}
	if input(req, &params, rw) != nil {
		return
	}
	result, err := s.service.FindPosts(params.Filter, params.Sort, params.Limit, params.Start)
	output(result, err, rw)
}

func (s *apiService) bulkSetVisible(rw http.ResponseWriter, req *http.Request) {
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

func (s *apiService) bulkUpdatePostConfigs(rw http.ResponseWriter, req *http.Request) {
	var params struct {
		PostIds []service.PostId
		Config  service.CommentConfig
	}
	if input(req, &params, rw) != nil {
		return
	}
	err := s.service.BulkUpdatePostConfigs(params.PostIds, params.Config)
	output("Success", err, rw)
}
