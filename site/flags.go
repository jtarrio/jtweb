package site

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

var flagConfigFile = flag.String("config_file", "", "The name of the file containing the site's configuration.")
var flagTemplatePath = flag.String("template_path", "", "The full pathname where the templates are located.")
var flagInputPath = flag.String("input_path", "", "The full pathname where the input files are located.")
var flagOutputPath = flag.String("output_path", "", "The full pathname where the rendered HTML files will be output.")
var flagWebroot = flag.String("webroot", "", "The URI where the generated content will live.")
var flagSiteName = flag.String("site_name", "", "The site's name.")
var flagSiteURI = flag.String("site_uri", "", "The site's URI.")
var flagHideUntranslated = flag.Bool("hide_untranslated", false, "Hide pages only available in a different language.")
var flagWebrootLanguages = flag.String("webroot_languages", "", "Per-language webroots, in lang=root[;lang=root...] format.")
var flagSiteNameLanguages = flag.String("site_name_languages", "", "Per-language site names, in lang=name[;lang=name...] format.")
var flagSiteURILanguages = flag.String("site_uri_languages", "", "Per-language site URIs, in lang=uri[;lang=uri...] format.")
var flagAuthorName = flag.String("author_name", "", "The default author's name.")
var flagAuthorURI = flag.String("author_uri", "", "The default author's website URI.")
var flagCurrentTime = flag.String("current_time", "", "The time to use instead of the current time.")

func parseByLanguage(cfg string) (map[string]string, error) {
	out := make(map[string]string)
	parts := strings.Split(cfg, ";")
	for _, part := range parts {
		lang, value, found := strings.Cut(part, "=")
		if !found {
			return nil, fmt.Errorf("syntax error in %s", part)
		}
		out[lang] = value
	}
	return out, nil
}

func FromFlags() (*Config, error) {
	var cfg *Config

	if *flagConfigFile != "" {
		file, err := ioutil.ReadFile(*flagConfigFile)
		if err != nil {
			return nil, err
		}
		cfg, err = ParseConfig(file)
		if err != nil {
			return nil, err
		}
	} else {
		cfg = &Config{}
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
	if *flagHideUntranslated {
		cfg.HideUntranslated = true
	}
	if *flagWebrootLanguages != "" {
		parsed, err := parseByLanguage(*flagWebrootLanguages)
		if err != nil {
			return nil, err
		}
		cfg.WebRootLanguages = parsed
	}
	if *flagSiteNameLanguages != "" {
		parsed, err := parseByLanguage(*flagSiteNameLanguages)
		if err != nil {
			return nil, err
		}
		cfg.SiteNameLanguages = parsed
	}
	if *flagSiteURILanguages != "" {
		parsed, err := parseByLanguage(*flagSiteURILanguages)
		if err != nil {
			return nil, err
		}
		cfg.SiteURILanguages = parsed
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
			return nil, err
		}
		cfg.CurrentTime = parsed
	} else {
		cfg.CurrentTime = time.Now()
	}

	err := cfg.Normalize()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
