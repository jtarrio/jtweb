package main

import (
	"flag"
	"io/ioutil"
	"time"

	"jacobo.tarrio.org/jtweb/site"
)

var flagConfigFile = flag.String("config_file", "", "The name of the file containing the site's configuration.")
var flagTemplatePath = flag.String("template_path", "", "The full pathname where the templates are located.")
var flagInputPath = flag.String("input_path", "", "The full pathname where the input files are located.")
var flagOutputPath = flag.String("output_path", "", "The full pathname where the rendered HTML files will be output.")
var flagWebroot = flag.String("webroot", "", "The URI where the generated content will live.")
var flagSiteName = flag.String("site_name", "", "The site's name.")
var flagSiteURI = flag.String("site_uri", "", "The site's URI.")
var flagAuthorName = flag.String("author_name", "", "The default author's name.")
var flagAuthorURI = flag.String("author_uri", "", "The default author's website URI.")
var flagCurrentTime = flag.String("current_time", "", "The time to use instead of the current time.")

func main() {
	flag.Parse()

	var cfg *site.Config

	if *flagConfigFile != "" {
		file, err := ioutil.ReadFile(*flagConfigFile)
		if err != nil {
			panic(err)
		}
		cfg, err = site.ParseConfig(file)
		if err != nil {
			panic(err)
		}
	} else {
		cfg = &site.Config{}
	}

	if *flagTemplatePath != "" {
		cfg.TemplatePath = *flagTemplatePath
	}
	if *flagInputPath != "" {
		cfg.InputPath = *flagInputPath
	}
	if *flagOutputPath != "" {
		cfg.OutputPath = *flagOutputPath
	}
	if *flagWebroot != "" {
		cfg.WebRoot = *flagWebroot
	}
	if *flagSiteName != "" {
		cfg.SiteName = *flagSiteName
	}
	if *flagSiteURI != "" {
		cfg.SiteURI = *flagSiteURI
	}
	if *flagAuthorName != "" {
		cfg.AuthorName = *flagAuthorName
	}
	if *flagAuthorURI != "" {
		cfg.AuthorURI = *flagAuthorURI
	}
	if *flagCurrentTime != "" {
		parsed, err := time.Parse(time.RFC3339, *flagCurrentTime)
		if err != nil {
			panic(err)
		}
		cfg.CurrentTime = parsed
	} else {
		cfg.CurrentTime = time.Now()
	}

	err := cfg.Normalize()
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
