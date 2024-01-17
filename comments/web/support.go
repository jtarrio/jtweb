package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/time/rate"
)

type webService struct {
	handlers     map[handlerPath]http.HandlerFunc
	adminChecker *AdminChecker
}

type handlerPath struct {
	prefix string
	method string
	admin  bool
}

func userPost(path string) handlerPath {
	return handlerPath{prefix: path, method: http.MethodPost, admin: false}
}

func adminPost(path string) handlerPath {
	return handlerPath{prefix: path, method: http.MethodPost, admin: true}
}

func (s *webService) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
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

func (s *webService) isAdmin(req *http.Request) bool {
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
	log.Printf("Error: %s", err)
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
