package fromflags

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

type byLanguage map[string]string

func ByLanguageFlag(name string, usage string) *map[string]string {
	p := &byLanguage{}
	flag.Var(p, name, usage)
	return (*map[string]string)(p)
}

func (bl *byLanguage) String() string {
	var out string
	for lang, val := range *bl {
		if out != "" {
			out = out + ";"
		}
		out = out + lang + "=" + strconv.Quote(val)
	}
	return out
}

func (bl *byLanguage) Set(value string) error {
	eq := strings.IndexRune(value, '=')
	if eq == -1 {
		return fmt.Errorf("expected an equals sign to be present: %s", value)
	}
	(*bl)[value[0:eq]] = value[eq+1:]
	return nil
}
