package lib

import (
	"jacobo.tarrio.org/jtweb/comments"
	"jacobo.tarrio.org/jtweb/comments/service"
	"jacobo.tarrio.org/jtweb/site"
)

func OpComments() OpFn {
	return func(contents *site.RawContents) error {
		cfg := contents.Config.Comments()
		posts := &service.AvailablePosts{Posts: map[comments.PostId]service.CommentConfig{}}
		for name, page := range contents.Pages {
			cfg := service.CommentConfig{
				IsAvailable: page.Header.Comments.Enabled,
				IsWritable:  page.Header.Comments.Writable,
			}
			posts.Posts[comments.PostId(name)] = cfg
		}
		return cfg.Service().SetAvailablePosts(posts)
	}
}
