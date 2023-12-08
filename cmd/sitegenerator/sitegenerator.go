package main

import (
	"flag"

	"jacobo.tarrio.org/jtweb/config"
	"jacobo.tarrio.org/jtweb/site"
)

func main() {
	flag.Parse()

	cfg, err := config.GetConfig()
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
