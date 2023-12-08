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
var flagOutputPath = flag.String("output_path", "", "The full pathname where the rendered HTML files will be output.")
var flagWebroot = flag.String("webroot", "", "The URI where the generated content will live.")
var flagPublishUntil = TimeFlag("publish_until", "Publish all posts older than the given date/time.")

type yamlConfig struct {
	Files struct {
		Templates string
		Content   string
	}
	Site struct {
		Webroot    string
		Name       string
		Uri        string
		ByLanguage map[string]struct {
			Webroot string
			Name    string
			Uri     string
		} `yaml:"by_language"`
	}
	Author struct {
		Name string
		Uri  string
	}
	Generator *struct {
		Output           string
		HideUntranslated bool       `yaml:"hide_untranslated"`
		PublishUntil     *time.Time `yaml:"publish_until"`
	}
}

type parsedConfig struct {
	files     fileConfig
	site      allSiteConfig
	author    authorConfig
	generator *generatorConfig
}

type fileConfig struct {
	templates io.File
	content   io.File
}

type allSiteConfig struct {
	defaultSite siteConfig
	byLanguage  map[languages.Language]siteConfig
}

type siteConfig struct {
	webroot string
	name    string
	uri     string
}
type authorConfig struct {
	name string
	uri  string
}
type generatorConfig struct {
	output           io.File
	hideUntranslated bool
	publishUntil     *time.Time
	now              time.Time
}

func readConfig(fileName string) (*yamlConfig, error) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var cfg = &yamlConfig{}
	err = yaml.UnmarshalStrict(file, cfg)
	return cfg, err
}

func convertConfig(cfg *yamlConfig) (config.Config, error) {
	var out = &parsedConfig{}
	if cfg.Files.Templates == "" {
		return nil, fmt.Errorf("the template path has not been set")
	}
	if cfg.Files.Content == "" {
		return nil, fmt.Errorf("the content path has not been set")
	}
	out.files = fileConfig{
		templates: io.OsFile(cfg.Files.Templates),
		content:   io.OsFile(cfg.Files.Content),
	}
	if cfg.Site.Webroot == "" {
		return nil, fmt.Errorf("the web root has not been set")
	}
	if cfg.Site.Name == "" {
		return nil, fmt.Errorf("the site name has not been set")
	}
	if cfg.Site.Uri == "" {
		cfg.Site.Uri = cfg.Site.Webroot
	}
	out.site.defaultSite = siteConfig{
		webroot: cfg.Site.Webroot,
		name:    cfg.Site.Name,
		uri:     cfg.Site.Uri,
	}
	for lang, siteCfg := range cfg.Site.ByLanguage {
		language, err := languages.FindByCode(lang)
		if err != nil {
			return nil, err
		}
		if siteCfg.Webroot == "" {
			siteCfg.Webroot = cfg.Site.Webroot
		}
		if siteCfg.Name == "" {
			siteCfg.Name = cfg.Site.Name
		}
		if siteCfg.Uri == "" {
			siteCfg.Uri = cfg.Site.Uri
		}
		out.site.byLanguage = make(map[languages.Language]siteConfig)
		out.site.byLanguage[language] = siteConfig{
			webroot: siteCfg.Webroot,
			name:    siteCfg.Name,
			uri:     siteCfg.Uri,
		}
	}
	if cfg.Author.Name == "" {
		return nil, fmt.Errorf("the author's name has not been set")
	}
	if cfg.Author.Uri == "" {
		cfg.Author.Uri = cfg.Site.Uri
	}
	out.author = authorConfig{
		name: cfg.Author.Name,
		uri:  cfg.Author.Uri,
	}
	if cfg.Generator != nil {
		if cfg.Generator.Output == "" {
			return nil, fmt.Errorf("the output path has not been set")
		}
		out.generator = &generatorConfig{
			output:           io.OsFile(cfg.Generator.Output),
			hideUntranslated: cfg.Generator.HideUntranslated,
			publishUntil:     cfg.Generator.PublishUntil,
			now:              time.Now(),
		}
	}

	return out, nil
}

func ReadConfig(fileName string) (config.Config, error) {
	cfg, err := readConfig(fileName)
	if err != nil {
		return nil, err
	}

	return convertConfig(cfg)
}

func GetConfig() (config.Config, error) {
	if *flagConfigFile == "" {
		return nil, fmt.Errorf("the --config_file flag has not been specified")
	}

	cfg, err := readConfig(*flagConfigFile)
	if err != nil {
		return nil, err
	}

	if *flagWebroot != "" {
		cfg.Site.Webroot = *flagWebroot
		for lang, langCfg := range cfg.Site.ByLanguage {
			langCfg.Webroot = *flagWebroot
			cfg.Site.ByLanguage[lang] = langCfg
		}
	}
	if *flagOutputPath != "" {
		if cfg.Generator == nil {
			return nil, fmt.Errorf("the --output flag has been specified but no 'generator' field is present in the configuration")
		}
		cfg.Generator.Output = *flagOutputPath
	}
	if !flagPublishUntil.IsZero() && cfg.Generator != nil {
		cfg.Generator.PublishUntil = flagPublishUntil
	}

	return convertConfig(cfg)
}

func (c *parsedConfig) Files() config.FileConfig {
	return &c.files
}

func (fc *fileConfig) Templates() io.File {
	return fc.templates
}

func (fc *fileConfig) Content() io.File {
	return fc.content
}

type foundSiteConfig struct {
	cfg  *siteConfig
	lang languages.Language
}

func (c *parsedConfig) Site(lang languages.Language) config.SiteConfig {
	cfg, ok := c.site.byLanguage[lang]
	if !ok {
		cfg = c.site.defaultSite
	}
	return &foundSiteConfig{&cfg, lang}
}

func (sc *foundSiteConfig) WebRoot() string {
	return sc.cfg.webroot
}

func (sc *foundSiteConfig) Name() string {
	return sc.cfg.name
}

func (sc *foundSiteConfig) Uri() string {
	return sc.cfg.uri
}

func (sc *foundSiteConfig) Language() languages.Language {
	return sc.lang
}

func (c *parsedConfig) Author() config.AuthorConfig {
	return &c.author
}

func (ac *authorConfig) Name() string {
	return ac.name
}

func (ac *authorConfig) Uri() string {
	return ac.uri
}

func (c *parsedConfig) Generator() config.GeneratorConfig {
	return c.generator
}

func (gc *generatorConfig) Output() io.File {
	return gc.output
}

func (gc *generatorConfig) HideUntranslated() bool {
	return gc.hideUntranslated
}

func (gc *generatorConfig) PublishUntil() time.Time {
	if gc.publishUntil == nil {
		return gc.now
	}
	return *gc.publishUntil
}
