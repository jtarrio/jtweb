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
	List(id PostId, seeDrafts bool) (*CommentList, error)
	Add(comment *NewComment) (*Comment, error)
	SetAvailablePosts(posts *AvailablePosts) error
}

type CommentList struct {
	PostId      PostId
	IsAvailable bool
	IsWritable  bool
	List        []*Comment
}

type Comment struct {
	Id      CommentId
	Visible bool
	Author  string
	When    time.Time
	Text    Html
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

func NewCommentsService(engine engine.Engine, options ...CommentsServiceOptions) CommentsService {
	service := &commentsServiceImpl{engine: engine, renderer: NewEscapeRenderer(), defaultVisible: true}
	for _, option := range options {
		option(service)
	}
	return service
}

type CommentsServiceOptions func(*commentsServiceImpl)

func PostAsDraft() CommentsServiceOptions {
	return func(s *commentsServiceImpl) {
		s.defaultVisible = false
	}
}

func WithRenderer(renderer Renderer) CommentsServiceOptions {
	return func(s *commentsServiceImpl) {
		s.renderer = renderer
	}
}

type commentsServiceImpl struct {
	engine         engine.Engine
	renderer       Renderer
	defaultVisible bool
}

func (s *commentsServiceImpl) List(id PostId, seeDrafts bool) (*CommentList, error) {
	cfg, err := s.engine.GetConfig(id)
	if err != nil {
		return nil, err
	}
	out := &CommentList{
		PostId:      id,
		IsAvailable: cfg.State != engine.CommentsDisabled,
		IsWritable:  cfg.State == engine.CommentsEnabled,
		List:        []*Comment{}}
	if !out.IsAvailable {
		return out, nil
	}
	list, err := s.engine.List(id, seeDrafts)
	if err != nil {
		return nil, err
	}
	for _, comment := range list {
		cmt, err := s.parseComment(&comment)
		if err != nil {
			return nil, err
		}
		out.List = append(out.List, cmt)
	}
	return out, nil
}

func (s *commentsServiceImpl) parseComment(comment *engine.Comment) (*Comment, error) {
	html, err := s.renderer.Render(comment.Text)
	if err != nil {
		return nil, err
	}
	cmt := &Comment{
		Id:      comment.CommentId,
		Visible: comment.Visible,
		Author:  comment.Author,
		When:    comment.When,
		Text:    html,
	}
	return cmt, nil
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
		PostId:  comment.PostId,
		Visible: s.defaultVisible,
		Author:  comment.Author,
		When:    comment.When,
		Text:    comment.Text,
	})
	if err != nil {
		return nil, err
	}
	return s.parseComment(nc)
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
