package yamlconfig

import (
	"time"

	comments "jacobo.tarrio.org/jtweb/comments/service"
	"jacobo.tarrio.org/jtweb/config"
	"jacobo.tarrio.org/jtweb/email"
	"jacobo.tarrio.org/jtweb/io"
	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/page"
)

type parsedConfig struct {
	files       fileConfig
	site        allSiteConfig
	author      authorConfig
	generator   *generatorConfig
	mailers     []config.MailerConfig
	comments    *commentsConfig
	dateFilters dateFilterConfig
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
	skipOperation    bool
}

type mailerConfig struct {
	name          string
	language      languages.Language
	subjectPrefix string
	engine        email.Engine
	skipOperation bool
}

type commentsConfig struct {
	defaultConfig *page.CommentConfig
	jsUri         string
	service       comments.CommentsService
	adminPassword string
	skipOperation bool
}

type dateFilterConfig struct {
	now      time.Time
	generate dateFilter
	mail     dateFilter
}

type dateFilter struct {
	notBefore *time.Time
	notAfter  *time.Time
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

func (gc *generatorConfig) SkipOperation() bool {
	return gc.skipOperation
}

func (gc *generatorConfig) Present() bool {
	return gc != nil
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

func (mc *mailerConfig) SkipOperation() bool {
	return mc.skipOperation
}

func (c *parsedConfig) Comments() config.CommentsConfig {
	return c.comments
}

func (cc *commentsConfig) DefaultConfig() *page.CommentConfig {
	if cc == nil {
		return &page.CommentConfig{Enabled: false, Writable: false}
	}
	return cc.defaultConfig
}

func (cc *commentsConfig) JsUri() string {
	return cc.jsUri
}

func (cc *commentsConfig) Service() comments.CommentsService {
	return cc.service
}

func (cc *commentsConfig) AdminPassword() string {
	return cc.adminPassword
}

func (cc *commentsConfig) SkipOperation() bool {
	return cc.skipOperation
}

func (cc *commentsConfig) Present() bool {
	return cc != nil
}

func (c *parsedConfig) DateFilters() config.DateFilterConfig {
	return &c.dateFilters
}

func (dfc *dateFilterConfig) Now() time.Time {
	return dfc.now
}

func (dfc *dateFilterConfig) Generate() config.DateFilter {
	return &dfc.generate
}

func (dfc *dateFilterConfig) Mail() config.DateFilter {
	return &dfc.mail
}

func (df *dateFilter) NotBefore() *time.Time {
	return df.notBefore
}

func (df *dateFilter) NotAfter() *time.Time {
	return df.notAfter
}
