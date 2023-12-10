package yamlconfig

import (
	"time"

	"jacobo.tarrio.org/jtweb/config"
	"jacobo.tarrio.org/jtweb/email"
	"jacobo.tarrio.org/jtweb/io"
	"jacobo.tarrio.org/jtweb/languages"
)

type parsedConfig struct {
	files     fileConfig
	site      allSiteConfig
	author    authorConfig
	generator *generatorConfig
	mailers   []config.MailerConfig
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

type mailerConfig struct {
	name          string
	language      languages.Language
	subjectPrefix string
	sendAfter     time.Time
	engine        email.Engine
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

func (c *parsedConfig) Mailers() []config.MailerConfig {
	return c.mailers
}

func (mc *mailerConfig) Name() string {
	return mc.name
}

func (mc *mailerConfig) Language() languages.Language {
	return mc.language
}

func (mc *mailerConfig) Engine() email.Engine {
	return mc.engine
}

func (mc *mailerConfig) SubjectPrefix() string {
	return mc.subjectPrefix
}

func (mc *mailerConfig) SendAfter() time.Time {
	return mc.sendAfter
}
