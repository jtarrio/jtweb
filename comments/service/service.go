package service

import (
	"fmt"
	"html"
	"time"

	"jacobo.tarrio.org/jtweb/comments"
	"jacobo.tarrio.org/jtweb/comments/engine"
)

type PostId = comments.PostId
type CommentId = comments.CommentId
type Markdown = comments.Markdown
type Html = comments.Html

type CommentsService interface {
	Get(id PostId) (*CommentList, error)
	Add(comment *NewComment) (*Comment, error)
	SetAvailablePosts(posts *AvailablePosts) error
}

type CommentList struct {
	PostId      PostId
	IsAvailable bool
	IsWritable  bool
	List        []Comment
}

type Comment struct {
	Id     CommentId
	Author string
	When   time.Time
	Text   Html
}

type NewComment struct {
	PostId PostId
	Author string
	When   time.Time
	Text   Markdown
}

type AvailablePosts struct {
	Posts map[comments.PostId]CommentConfig
}

type CommentConfig struct {
	IsAvailable bool
	IsWritable  bool
}

func NewCommentsService(engine engine.Engine) CommentsService {
	return &commentsServiceImpl{engine: engine}
}

type commentsServiceImpl struct {
	engine engine.Engine
}

func (s *commentsServiceImpl) Get(id PostId) (*CommentList, error) {
	cfg, err := s.engine.GetConfig(id)
	if err != nil {
		return nil, err
	}
	out := &CommentList{
		PostId:      id,
		IsAvailable: cfg.State != engine.CommentsDisabled,
		IsWritable:  cfg.State == engine.CommentsEnabled,
		List:        []Comment{}}
	if !out.IsAvailable {
		return out, nil
	}
	list, err := s.engine.Load(id)
	if err != nil {
		return nil, err
	}
	for _, comment := range list {
		out.List = append(out.List, parseComment(&comment))
	}
	return out, nil
}

func parseComment(comment *engine.Comment) Comment {
	return Comment{
		Id:     comment.CommentId,
		Author: comment.Author,
		When:   comment.When,
		Text:   Html(html.EscapeString(string(comment.Text))),
	}
}

func (s *commentsServiceImpl) Add(comment *NewComment) (*Comment, error) {
	cfg, err := s.engine.GetConfig(comment.PostId)
	if err != nil {
		return nil, err
	}
	if cfg.State != engine.CommentsEnabled {
		return nil, fmt.Errorf("comments are closed for post [%s]", comment.PostId)
	}
	nc, err := s.engine.Add(&engine.NewComment{
		PostId: comment.PostId,
		Author: comment.Author,
		When:   comment.When,
		Text:   comment.Text,
	})
	if err != nil {
		return nil, err
	}
	parsed := parseComment(nc)
	return &parsed, nil
}

func (s *commentsServiceImpl) SetAvailablePosts(posts *AvailablePosts) error {
	cfg := &engine.BulkConfig{Configs: make([]engine.Config, 0)}
	for id, postCfg := range posts.Posts {
		state := engine.CommentsDisabled
		if postCfg.IsWritable {
			state = engine.CommentsEnabled
		} else if postCfg.IsAvailable {
			state = engine.CommentsClosed
		}
		cfg.Configs = append(cfg.Configs, engine.Config{PostId: id, State: state})
	}
	return s.engine.BulkSetConfig(cfg)
}
