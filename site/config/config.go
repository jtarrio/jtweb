package config

import (
	"time"

	"jacobo.tarrio.org/jtweb/site/io"
)

type Config interface {
	GetTemplateBase() io.File
	GetInputBase() io.File
	GetOutputBase() io.File
	GetWebRoot(lang string) string
	GetSiteName(lang string) string
	GetSiteURI(lang string) string
	GetAuthorName() string
	GetAuthorURI() string
	GetHideUntranslated() bool
	GetPublishUntil() time.Time
}
