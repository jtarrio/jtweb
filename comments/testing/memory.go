package testing

import (
	"fmt"

	"jacobo.tarrio.org/jtweb/comments"
	"jacobo.tarrio.org/jtweb/comments/engine"
)

type MemoryEngine struct {
	posts        map[comments.PostId]bool
	config       map[comments.PostId]*engine.Config
	comments     map[comments.PostId][]engine.Comment
	defaultState engine.CommentState
	lastId       uint64
}

func NewMemoryEngine() *MemoryEngine {
	return &MemoryEngine{
		posts:        map[comments.PostId]bool{},
		comments:     map[comments.PostId][]engine.Comment{},
		defaultState: engine.CommentsDisabled,
		lastId:       1000,
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

func (e *MemoryEngine) GetConfig(postId comments.PostId) (*engine.Config, error) {
	err := e.CheckPost(postId)
	if err != nil {
		return nil, err
	}
	c, ok := e.config[postId]
	if !ok {
		return &engine.Config{PostId: postId, State: e.defaultState}, nil
	}
	return c, nil
}

func (e *MemoryEngine) SetConfig(newConfig, oldConfig *engine.Config) error {
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

func (e *MemoryEngine) Load(postId comments.PostId) ([]engine.Comment, error) {
	err := e.CheckPost(postId)
	if err != nil {
		return nil, err
	}
	c, ok := e.comments[postId]
	if !ok {
		return []engine.Comment{}, nil
	}
	return c, nil
}

func (e *MemoryEngine) Add(comment *engine.NewComment) (*engine.Comment, error) {
	err := e.CheckPost(comment.PostId)
	if err != nil {
		return nil, err
	}
	nc := engine.Comment{
		PostId:    comment.PostId,
		CommentId: e.nextId(),
		Author:    comment.Author,
		When:      comment.When,
		Text:      comment.Text,
	}
	e.comments[comment.PostId] = append(e.comments[comment.PostId], nc)
	return &nc, nil
}
