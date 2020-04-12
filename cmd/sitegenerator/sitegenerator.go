package main

import (
	"flag"
	"io/ioutil"
	"jtweb/site"

	"gopkg.in/yaml.v2"
)

var flagConfigFile = flag.String("config_file", "", "The name of the file containing the site's configuration.")
var flagTemplatePath = flag.String("template_path", "", "The full pathname where the templates are located.")
var flagInputPath = flag.String("input_path", "", "The full pathname where the input files are located.")
var flagOutputPath = flag.String("output_path", "", "The full pathname where the rendered HTML files will be output.")
var flagWebroot = flag.String("webroot", "", "The URI where the generated content will live.")
var flagAuthorName = flag.String("author_name", "", "The default author's name.")
var flagAuthorURI = flag.String("author_uri", "", "The default author's website URI.")

func main() {
	flag.Parse()

	cfg := site.Config{}

	if *flagConfigFile != "" {
		file, err := ioutil.ReadFile(*flagConfigFile)
		if err != nil {
			panic(err)
		}
		err = yaml.UnmarshalStrict(file, &cfg)
		if err != nil {
			panic(err)
		}
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
	if *flagAuthorName != "" {
		cfg.AuthorName = *flagAuthorName
	}
	if *flagAuthorURI != "" {
		cfg.AuthorURI = *flagAuthorURI
	}

	if cfg.TemplatePath == "" {
		panic("The template path has not been set.")
	}
	if cfg.InputPath == "" {
		panic("The input path has not been set.")
	}
	if cfg.OutputPath == "" {
		panic("The output path has not been set.")
	}
	if cfg.WebRoot == "" {
		panic("The web root has not been set.")
	}
	if cfg.AuthorName == "" {
		panic("The default author's name has not been set.")
	}
	if cfg.AuthorURI == "" {
		cfg.AuthorURI = cfg.WebRoot
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
