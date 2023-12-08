package config

import (
	"time"

	"jacobo.tarrio.org/jtweb/io"
	"jacobo.tarrio.org/jtweb/languages"
)

type Config interface {
	Files() FileConfig
	Site(lang languages.Language) SiteConfig
	Author() AuthorConfig
	Generator() GeneratorConfig
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
	PublishUntil() time.Time
}
