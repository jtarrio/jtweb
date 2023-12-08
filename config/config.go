package config

import (
	"time"

	"jacobo.tarrio.org/jtweb/io"
	"jacobo.tarrio.org/jtweb/languages"
)

type Config interface {
	GetTemplateBase() io.File
	GetInputBase() io.File
	GetOutputBase() io.File
	GetWebRoot(lang languages.Language) string
	GetSiteName(lang languages.Language) string
	GetSiteURI(lang languages.Language) string
	GetAuthorName() string
	GetAuthorURI() string
	GetHideUntranslated() bool
	GetPublishUntil() time.Time
}
