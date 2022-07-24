package main

import (
	"flag"

	"jacobo.tarrio.org/jtweb/site"
)

func main() {
	flag.Parse()

	cfg, err := site.FromFlags()
	if err != nil {
		panic(err)
	}
	content, err := cfg.Read()
	if err != nil {
		panic(err)
	}
	err = content.Write()
	if err != nil {
		panic(err)
	}
}
