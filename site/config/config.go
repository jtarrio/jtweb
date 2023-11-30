package config

import (
	"time"

	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/site/io"
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
