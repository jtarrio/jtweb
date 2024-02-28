package engine

import (
	"context"
	"time"

	"jacobo.tarrio.org/jtweb/comments"
)

type PostId = comments.PostId
type CommentId = comments.CommentId
type Markdown = comments.Markdown

type CommentState int

const (
	CommentsDisabled = CommentState(0)
	CommentsClosed   = CommentState(1)
	CommentsEnabled  = CommentState(2)
)

type Config struct {
	PostId PostId
	State  CommentState
}

type Comment struct {
	PostId    PostId
	CommentId CommentId
	Visible   bool
	Author    string
	When      time.Time
	Text      Markdown
}

type NewComment struct {
	PostId  PostId
	Visible bool
	Author  string
	When    time.Time
	Text    Markdown
}

type BulkConfig struct {
	Configs []Config
}

type CommentFilter struct {
	Visible *bool
}

type PostFilter struct {
	CommentsReadable *bool
	CommentsWritable *bool
}

type Sort int

const (
	SortNewestFirst = Sort(iota)
)

type Engine interface {
	GetConfig(ctx context.Context, postId PostId) (*Config, error)
	SetConfig(ctx context.Context, newConfig, oldConfig *Config) error
	SetAllPostConfigs(ctx context.Context, cfg *BulkConfig) error
	BulkUpdatePostConfigs(ctx context.Context, cfg *BulkConfig) error
	List(ctx context.Context, postId PostId, seeDrafts bool) ([]*Comment, error)
	Add(ctx context.Context, comment *NewComment) (*Comment, error)
	FindComments(ctx context.Context, filter CommentFilter, sort Sort, limit int, start int) ([]*Comment, error)
	DeleteComments(ctx context.Context, ids map[PostId][]*CommentId) error
	FindPosts(ctx context.Context, filter PostFilter, sort Sort, limit int, start int) ([]*Config, error)
	BulkSetVisible(ctx context.Context, ids map[PostId][]*CommentId, visible bool) error
}
