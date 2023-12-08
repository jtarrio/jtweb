package fromflags

import (
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
	"jacobo.tarrio.org/jtweb/config"
	"jacobo.tarrio.org/jtweb/io"
	"jacobo.tarrio.org/jtweb/languages"
)

var flagConfigFile = flag.String("config_file", "", "The name of the file containing the site's configuration.")
var flagTemplatePath = flag.String("template_path", "", "The full pathname where the templates are located.")
var flagInputPath = flag.String("input_path", "", "The full pathname where the input files are located.")
var flagOutputPath = flag.String("output_path", "", "The full pathname where the rendered HTML files will be output.")
var flagWebroot = flag.String("webroot", "", "The URI where the generated content will live.")
var flagSiteName = flag.String("site_name", "", "The site's name.")
var flagSiteURI = flag.String("site_uri", "", "The site's URI.")
var flagHideUntranslated = flag.Bool("hide_untranslated", false, "Hide pages only available in a different language.")
var flagWebrootLanguages = ByLanguageFlag("webroot_languages", "Per-language webroots in lang=root format, repeated.")
var flagSiteNameLanguages = ByLanguageFlag("site_name_languages", "Per-language site names in lang=name format, repeated.")
var flagSiteURILanguages = ByLanguageFlag("site_uri_languages", "Per-language site URIs in lang=uri format, repeated.")
var flagAuthorName = flag.String("author_name", "", "The default author's name.")
var flagAuthorURI = flag.String("author_uri", "", "The default author's website URI.")
var flagPublishUntil = TimeFlag("publish_until", "Publish all posts older than the given date/time.")

type configFromFlags struct {
	TemplatePath      string            `yaml:"template_path"`
	InputPath         string            `yaml:"input_path"`
	OutputPath        string            `yaml:"output_path"`
	WebRoot           string            `yaml:"webroot"`
	SiteName          string            `yaml:"site_name"`
	SiteURI           string            `yaml:"site_uri"`
	HideUntranslated  bool              `yaml:"hide_untranslated"`
	WebRootLanguages  map[string]string `yaml:"webroot_languages"`
	SiteNameLanguages map[string]string `yaml:"site_name_languages"`
	SiteURILanguages  map[string]string `yaml:"site_uri_languages"`
	AuthorName        string            `yaml:"author_name"`
	AuthorURI         string            `yaml:"author_uri"`
	PublishUntil      time.Time         `yaml:"publish_until"`
}

func newConfig() *configFromFlags {
	return &configFromFlags{
		PublishUntil: time.Now(),
	}
}

func (c *configFromFlags) mergeYaml(b []byte) error {
	return yaml.UnmarshalStrict(b, c)
}

func (c *configFromFlags) mergeFlags() error {
	if *flagTemplatePath != "" {
		c.TemplatePath = *flagTemplatePath
	}
	if *flagInputPath != "" {
		c.InputPath = *flagInputPath
	}
	if *flagOutputPath != "" {
		c.OutputPath = *flagOutputPath
	}
	if *flagWebroot != "" {
		c.WebRoot = *flagWebroot
	}
	if *flagSiteName != "" {
		c.SiteName = *flagSiteName
	}
	if *flagSiteURI != "" {
		c.SiteURI = *flagSiteURI
	}
	if *flagHideUntranslated {
		c.HideUntranslated = true
	}
	for k, v := range *flagWebrootLanguages {
		c.WebRootLanguages[k] = v
	}
	for k, v := range *flagSiteNameLanguages {
		c.SiteNameLanguages[k] = v
	}
	for k, v := range *flagSiteURILanguages {
		c.SiteURILanguages[k] = v
	}
	if *flagAuthorName != "" {
		c.AuthorName = *flagAuthorName
	}
	if *flagAuthorURI != "" {
		c.AuthorURI = *flagAuthorURI
	}
	if !flagPublishUntil.IsZero() {
		c.PublishUntil = *flagPublishUntil
	}
	return nil
}

// Normalize checks that the configuration is valid and fills any missing optional values.
func (c *configFromFlags) normalize() error {
	if c.TemplatePath == "" {
		return fmt.Errorf("the template path has not been set")
	}
	if c.InputPath == "" {
		return fmt.Errorf("the input path has not been set")
	}
	if c.OutputPath == "" {
		return fmt.Errorf("the output path has not been set")
	}
	if c.WebRoot == "" {
		return fmt.Errorf("the web root has not been set")
	}
	if c.SiteName == "" {
		return fmt.Errorf("the site name has not been set")
	}
	if c.AuthorName == "" {
		return fmt.Errorf("the default author's name has not been set")
	}
	if c.SiteURI == "" {
		c.SiteURI = c.WebRoot
	}
	if c.AuthorURI == "" {
		c.AuthorURI = c.SiteURI
	}
	if c.PublishUntil.IsZero() {
		c.PublishUntil = time.Now()
	}
	return nil
}

func GetConfig() (config.Config, error) {
	cfg := newConfig()

	if *flagConfigFile != "" {
		file, err := ioutil.ReadFile(*flagConfigFile)
		if err != nil {
			return nil, err
		}
		err = cfg.mergeYaml(file)
		if err != nil {
			return nil, err
		}
	}

	err := cfg.mergeFlags()
	if err != nil {
		return nil, err
	}

	err = cfg.normalize()
	return cfg, err
}

type fileConfig struct {
	cfg *configFromFlags
}

func (c *configFromFlags) Files() config.FileConfig {
	return &fileConfig{c}
}

func (fc *fileConfig) Templates() io.File {
	return io.OsFile(fc.cfg.TemplatePath)
}

func (fc *fileConfig) Input() io.File {
	return io.OsFile(fc.cfg.InputPath)
}

type siteConfig struct {
	cfg  *configFromFlags
	lang languages.Language
}

func (c *configFromFlags) Site(lang languages.Language) config.SiteConfig {
	return &siteConfig{c, lang}
}

func (sc *siteConfig) WebRoot() string {
	val, ok := sc.cfg.WebRootLanguages[sc.lang.Code()]
	if ok {
		return val
	}
	return sc.cfg.WebRoot
}

func (sc *siteConfig) Name() string {
	val, ok := sc.cfg.SiteNameLanguages[sc.lang.Code()]
	if ok {
		return val
	}
	return sc.cfg.SiteName
}

func (sc *siteConfig) Uri() string {
	val, ok := sc.cfg.SiteURILanguages[sc.lang.Code()]
	if ok {
		return val
	}
	return sc.cfg.SiteURI
}

func (sc *siteConfig) Language() languages.Language {
	return sc.lang
}

type authorConfig struct {
	cfg *configFromFlags
}

func (c *configFromFlags) Author() config.AuthorConfig {
	return &authorConfig{c}
}

func (ac *authorConfig) Name() string {
	return ac.cfg.AuthorName
}

func (ac *authorConfig) Uri() string {
	return ac.cfg.AuthorURI
}

type generatorConfig struct {
	cfg *configFromFlags
}

func (c *configFromFlags) Generator() config.GeneratorConfig {
	return &generatorConfig{c}
}

func (gc *generatorConfig) Output() io.File {
	return io.OsFile(gc.cfg.OutputPath)
}

func (gc *generatorConfig) HideUntranslated() bool {
	return gc.cfg.HideUntranslated
}

func (gc *generatorConfig) PublishUntil() time.Time {
	return gc.cfg.PublishUntil
}
