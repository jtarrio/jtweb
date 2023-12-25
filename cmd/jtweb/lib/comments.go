package lib

import (
	"jacobo.tarrio.org/jtweb/comments"
	"jacobo.tarrio.org/jtweb/comments/service"
	"jacobo.tarrio.org/jtweb/site"
)

func OpComments() OpFn {
	return func(contents *site.RawContents) error {
		cfg := contents.Config.Comments()
		defaultConfig := cfg.DefaultConfig()
		posts := &service.AvailablePosts{Posts: map[comments.PostId]service.CommentConfig{}}
		for name := range contents.Pages {
			cfg := service.CommentConfig{
				IsAvailable: defaultConfig.Enabled,
				IsWritable:  defaultConfig.Writable,
			}
			posts.Posts[comments.PostId(name)] = cfg
		}
		return cfg.Service().SetAvailablePosts(posts)
	}
}
