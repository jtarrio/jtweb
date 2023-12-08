package main

import (
	"flag"

	"jacobo.tarrio.org/jtweb/config/fromflags"
	"jacobo.tarrio.org/jtweb/site"
)

func main() {
	flag.Parse()

	cfg, err := fromflags.GetConfig()
	if err != nil {
		panic(err)
	}
	content, err := site.Read(cfg)
	if err != nil {
		panic(err)
	}
	err = content.Write()
	if err != nil {
		panic(err)
	}
}
