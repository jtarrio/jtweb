package engine

import (
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

type Filter struct {
	Visible *bool
}

type Sort int

const (
	SortNewestFirst = Sort(iota)
)

type Engine interface {
	GetConfig(postId PostId) (*Config, error)
	SetConfig(newConfig, oldConfig *Config) error
	BulkSetConfig(cfg *BulkConfig) error
	List(postId PostId, seeDrafts bool) ([]*Comment, error)
	Add(comment *NewComment) (*Comment, error)
	Find(filter Filter, sort Sort, limit int, start int) ([]*Comment, error)
	BulkSetVisible(ids map[PostId][]*CommentId, visible bool) error
}
