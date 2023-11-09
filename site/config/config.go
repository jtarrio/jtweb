package config

import (
	"time"
)

type Config interface {
	GetTemplatePath() string
	GetInputPath() string
	GetOutputPath() string
	GetWebRoot(lang string) string
	GetSiteName(lang string) string
	GetSiteURI(lang string) string
	GetAuthorName() string
	GetAuthorURI() string
	GetHideUntranslated() bool
	GetPublishUntil() time.Time
}
