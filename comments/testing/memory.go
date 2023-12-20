package testing

import (
	"fmt"

	"jacobo.tarrio.org/jtweb/comments"
)

type MemoryEngine struct {
	posts    map[comments.PostId]bool
	config   map[comments.PostId]*comments.Config
	comments map[comments.PostId][]comments.Comment
	lastId   uint64
}

func NewMemoryEngine() *MemoryEngine {
	return &MemoryEngine{
		comments: map[comments.PostId][]comments.Comment{},
		lastId:   1000,
	}
}

func (e *MemoryEngine) nextId() comments.CommentId {
	e.lastId++
	return comments.CommentId(fmt.Sprint(e.lastId))
}

func (e *MemoryEngine) AddPost(postId comments.PostId) {
	e.posts[postId] = true
}

func (e *MemoryEngine) CheckPost(postId comments.PostId) error {
	_, ok := e.posts[postId]
	if !ok {
		return fmt.Errorf("unknown post [%s]", postId)
	}
	return nil
}

func (e *MemoryEngine) GetConfig(postId comments.PostId) (*comments.Config, error) {
	err := e.CheckPost(postId)
	if err != nil {
		return nil, err
	}
	c, ok := e.config[postId]
	if !ok {
		return &comments.Config{PostId: postId, Enabled: comments.DefaultState}, nil
	}
	return c, nil
}

func (e *MemoryEngine) SetConfig(newConfig, oldConfig *comments.Config) error {
	c, err := e.GetConfig(newConfig.PostId)
	if err != nil {
		return err
	}
	if c != oldConfig {
		return fmt.Errorf("differences exist in previous configuration for post [%s]", newConfig.PostId)
	}
	e.config[newConfig.PostId] = newConfig
	return nil
}

func (e *MemoryEngine) Load(postId comments.PostId) ([]comments.Comment, error) {
	err := e.CheckPost(postId)
	if err != nil {
		return nil, err
	}
	c, ok := e.comments[postId]
	if !ok {
		return []comments.Comment{}, nil
	}
	return c, nil
}

func (e *MemoryEngine) Add(comment *comments.NewComment) (*comments.Comment, error) {
	err := e.CheckPost(comment.PostId)
	if err != nil {
		return nil, err
	}
	nc := comments.Comment{
		PostId:    comment.PostId,
		CommentId: e.nextId(),
		Author:    comment.Author,
		When:      comment.When,
		Text:      comment.Text,
	}
	e.comments[comment.PostId] = append(e.comments[comment.PostId], nc)
	return &nc, nil
}
