package engine

import (
	"time"

	"jacobo.tarrio.org/jtweb/comments"
)

type PostId = comments.PostId
type CommentId = comments.CommentId
type Markdown = comments.Markdown

type commentsEnabledState int

const (
	CommentsDisabled = commentsEnabledState(0)
	CommentsClosed   = commentsEnabledState(1)
	CommentsEnabled  = commentsEnabledState(2)
)

type Config struct {
	PostId  PostId
	Enabled commentsEnabledState
}

type Comment struct {
	PostId    PostId
	CommentId CommentId
	Author    string
	When      time.Time
	Text      Markdown
}

type NewComment struct {
	PostId PostId
	Author string
	When   time.Time
	Text   Markdown
}

type Engine interface {
	GetConfig(postId PostId) (*Config, error)
	SetConfig(newConfig, oldConfig *Config) error
	Load(postId PostId) ([]Comment, error)
	Add(comment *NewComment) (*Comment, error)
}
