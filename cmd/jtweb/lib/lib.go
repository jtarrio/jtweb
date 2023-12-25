package lib

import (
	"time"

	"jacobo.tarrio.org/jtweb/site"
)

type OpFn func(content *site.RawContents) error

func getTimeOrDefault(when *time.Time, def time.Time) *time.Time {
	if when == nil {
		return &def
	}
	return when
}
