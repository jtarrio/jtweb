package config

import (
	"time"

	comments "jacobo.tarrio.org/jtweb/comments/service"
	"jacobo.tarrio.org/jtweb/email"
	"jacobo.tarrio.org/jtweb/io"
	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/page"
)

type Config interface {
	Files() FileConfig
	Site(lang languages.Language) SiteConfig
	Author() AuthorConfig
	Generator() GeneratorConfig
	Mailers() []MailerConfig
	Comments() CommentsConfig
	DateFilters() DateFilterConfig
}

type FileConfig interface {
	Templates() io.File
	Content() io.File
}

type SiteConfig interface {
	WebRoot() string
	Name() string
	Uri() string
	Language() languages.Language
}

type AuthorConfig interface {
	Name() string
	Uri() string
}

type GeneratorConfig interface {
	Output() io.File
	HideUntranslated() bool
	SkipOperation() bool
	Present() bool
}

type MailerConfig interface {
	Name() string
	Language() languages.Language
	Engine() email.Engine
	SubjectPrefix() string
	SkipOperation() bool
}

type CommentsConfig interface {
	DefaultConfig() *page.CommentConfig
	JsUri() string
	Service() comments.CommentsService
	AdminPassword() string
	SkipOperation() bool
	Present() bool
}

type DateFilterConfig interface {
	Now() time.Time
	Generate() DateFilter
	Mail() DateFilter
}

type DateFilter interface {
	NotBefore() *time.Time
	NotAfter() *time.Time
}
