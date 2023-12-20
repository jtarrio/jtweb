package service

import (
	"fmt"
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
}

type CommentList struct {
	IsAvailable bool
	IsWritable  bool
	PostId      PostId
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

func NewService(engine engine.Engine) CommentsService {
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
		IsAvailable: cfg.Enabled != engine.CommentsDisabled,
		IsWritable:  cfg.Enabled == engine.CommentsEnabled,
		PostId:      id}
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
		Text:   Html(comment.Text),
	}
}

func (s *commentsServiceImpl) Add(comment *NewComment) (*Comment, error) {
	cfg, err := s.engine.GetConfig(comment.PostId)
	if err != nil {
		return nil, err
	}
	if cfg.Enabled != engine.CommentsEnabled {
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
