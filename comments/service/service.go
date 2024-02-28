package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"jacobo.tarrio.org/jtweb/comments"
	"jacobo.tarrio.org/jtweb/comments/engine"
	"jacobo.tarrio.org/jtweb/comments/notification"
)

type PostId = comments.PostId
type CommentId = comments.CommentId
type Markdown = comments.Markdown
type Html = comments.Html

type CommentsService interface {
	List(ctx context.Context, id PostId, seeDrafts bool) (*CommentList, error)
	Add(ctx context.Context, comment *NewComment) (*Comment, error)
	Render(ctx context.Context, text Markdown) (Html, error)
	FindComments(ctx context.Context, filter CommentFilter, sort Sort, limit int, start int) (*FoundComments, error)
	DeleteComments(ctx context.Context, ids map[PostId][]*CommentId) error
	FindPosts(ctx context.Context, filter PostFilter, sort Sort, limit int, start int) (*FoundPosts, error)
	BulkSetVisible(ctx context.Context, ids map[PostId][]*CommentId, visible bool) error
	SetAvailablePosts(ctx context.Context, posts *AvailablePosts) error
	BulkUpdatePostConfigs(ctx context.Context, ids []PostId, config CommentConfig) error
}

type CommentList struct {
	PostId PostId
	Config CommentConfig
	List   []*Comment
}

type Comment struct {
	Id      CommentId
	Visible bool
	Author  string
	When    time.Time
	Text    Html
}

type RawComment = engine.Comment

type NewComment struct {
	PostId PostId
	Author string
	When   time.Time
	Text   Markdown
}

type CommentFilter = engine.CommentFilter
type PostFilter = engine.PostFilter
type Sort = engine.Sort

const SortNewestFirst = engine.SortNewestFirst

type FoundComments struct {
	List []*RawComment
	More bool
}

type FoundPosts struct {
	List []*FoundPost
	More bool
}

type FoundPost struct {
	PostId comments.PostId
	Config CommentConfig
}

type AvailablePosts struct {
	Posts map[comments.PostId]CommentConfig
}

type CommentConfig struct {
	IsReadable bool
	IsWritable bool
}

func NewCommentsService(engine engine.Engine, options ...CommentsServiceOptions) CommentsService {
	service := &commentsServiceImpl{
		engine:         engine,
		notifications:  notification.NullNotificationEngine(),
		renderer:       NewEscapeRenderer(),
		defaultVisible: true}
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

func WithNotificationEngine(notifications notification.NotificationEngine) CommentsServiceOptions {
	return func(s *commentsServiceImpl) {
		s.notifications = notifications
	}
}

type commentsServiceImpl struct {
	engine         engine.Engine
	notifications  notification.NotificationEngine
	renderer       Renderer
	defaultVisible bool
}

func commentConfigToState(cfg CommentConfig) engine.CommentState {
	if cfg.IsReadable {
		if cfg.IsWritable {
			return engine.CommentsEnabled
		}
		return engine.CommentsClosed
	}
	return engine.CommentsDisabled
}

func commentStateToConfig(state engine.CommentState) CommentConfig {
	return CommentConfig{
		IsWritable: state == engine.CommentsEnabled,
		IsReadable: state != engine.CommentsDisabled,
	}
}

func (s *commentsServiceImpl) List(ctx context.Context, id PostId, seeDrafts bool) (*CommentList, error) {
	cfg, err := s.engine.GetConfig(ctx, id)
	if err != nil {
		return nil, err
	}
	out := &CommentList{
		PostId: id,
		Config: commentStateToConfig(cfg.State),
		List:   []*Comment{}}
	if !out.Config.IsReadable {
		return out, nil
	}
	list, err := s.engine.List(ctx, id, seeDrafts)
	if err != nil {
		return nil, err
	}
	for _, comment := range list {
		cmt, err := s.parseComment(comment)
		if err != nil {
			return nil, err
		}
		out.List = append(out.List, cmt)
	}
	return out, nil
}

func (s *commentsServiceImpl) FindComments(ctx context.Context, filter CommentFilter, sort Sort, limit int, start int) (*FoundComments, error) {
	if start < 0 {
		start = 0
	}
	list, err := s.engine.FindComments(ctx, filter, sort, limit+1, start)
	if err != nil {
		return nil, err
	}
	last := limit
	if last > len(list) {
		last = len(list)
	}
	out := &FoundComments{
		List: list[0:last],
		More: len(list) > limit,
	}
	return out, nil
}

func (s *commentsServiceImpl) DeleteComments(ctx context.Context, ids map[PostId][]*CommentId) error {
	return s.engine.DeleteComments(ctx, ids)
}

func (s *commentsServiceImpl) FindPosts(ctx context.Context, filter PostFilter, sort Sort, limit int, start int) (*FoundPosts, error) {
	if start < 0 {
		start = 0
	}
	list, err := s.engine.FindPosts(ctx, filter, sort, limit+1, start)
	if err != nil {
		return nil, err
	}
	out := &FoundPosts{
		List: []*FoundPost{},
		More: len(list) > limit,
	}
	for i := 0; i < len(list) && i < limit; i++ {
		out.List = append(out.List, &FoundPost{
			PostId: list[i].PostId,
			Config: commentStateToConfig(list[i].State),
		})
	}
	return out, nil
}

func (s *commentsServiceImpl) BulkSetVisible(ctx context.Context, ids map[PostId][]*CommentId, visible bool) error {
	return s.engine.BulkSetVisible(ctx, ids, visible)
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

func (s *commentsServiceImpl) Add(ctx context.Context, comment *NewComment) (*Comment, error) {
	cfg, err := s.engine.GetConfig(ctx, comment.PostId)
	if err != nil {
		return nil, err
	}
	if cfg.State != engine.CommentsEnabled {
		return nil, fmt.Errorf("comments are closed for post [%s]", comment.PostId)
	}
	nc, err := s.engine.Add(ctx, &engine.NewComment{
		PostId:  comment.PostId,
		Visible: s.defaultVisible,
		Author:  comment.Author,
		When:    comment.When,
		Text:    comment.Text,
	})
	if err != nil {
		return nil, err
	}
	err = s.notifications.Notify(nc)
	if err != nil {
		log.Printf("error sending notification: %s", err)
	}
	return s.parseComment(nc)
}

func (s *commentsServiceImpl) Render(ctx context.Context, text Markdown) (Html, error) {
	return s.renderer.Render(text)
}

func (s *commentsServiceImpl) SetAvailablePosts(ctx context.Context, posts *AvailablePosts) error {
	cfg := &engine.BulkConfig{Configs: make([]engine.Config, 0)}
	for id, postCfg := range posts.Posts {
		cfg.Configs = append(cfg.Configs, engine.Config{PostId: id, State: commentConfigToState(postCfg)})
	}
	return s.engine.SetAllPostConfigs(ctx, cfg)
}

func (s *commentsServiceImpl) BulkUpdatePostConfigs(ctx context.Context, ids []PostId, config CommentConfig) error {
	cfg := &engine.BulkConfig{
		Configs: []engine.Config{},
	}
	state := commentConfigToState(config)
	for _, id := range ids {
		cfg.Configs = append(cfg.Configs, engine.Config{
			PostId: id,
			State:  state,
		})
	}
	return s.engine.BulkUpdatePostConfigs(ctx, cfg)
}
