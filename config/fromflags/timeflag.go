package fromflags

import (
	"flag"
	"time"
)

type timeFlag time.Time

func TimeFlag(name string, usage string) *time.Time {
	p := &timeFlag{}
	flag.Var(p, name, usage)
	return (*time.Time)(p)
}

func (t *timeFlag) String() string {
	return (*time.Time)(t).Format(time.RFC3339)
}

func (bl *timeFlag) Set(value string) error {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return err
	}
	*bl = timeFlag(parsed)
	return nil
}
