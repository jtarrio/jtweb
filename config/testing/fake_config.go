package testing

import (
	"time"

	"jacobo.tarrio.org/jtweb/config"
	"jacobo.tarrio.org/jtweb/io"
	"jacobo.tarrio.org/jtweb/io/testing"
	"jacobo.tarrio.org/jtweb/languages"
)

type FakeConfig struct {
	TemplateBase     io.File
	InputBase        io.File
	OutputBase       io.File
	WebRoots         map[string]string
	SiteNames        map[string]string
	SiteURIs         map[string]string
	AuthorName       string
	AuthorURI        string
	HideUntranslated bool
	PublishUntil     time.Time
}

func NewFakeConfig() *FakeConfig {
	return &FakeConfig{
		TemplateBase:     testing.NewMemoryFs(),
		InputBase:        testing.NewMemoryFs(),
		OutputBase:       testing.NewMemoryFs(),
		WebRoots:         map[string]string{"": "http://webroot"},
		SiteNames:        map[string]string{"": "Site Name"},
		SiteURIs:         map[string]string{"": "http://site"},
		AuthorName:       "Author",
		AuthorURI:        "http://author",
		HideUntranslated: false,
		PublishUntil:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

type fileConfig struct {
	cfg *FakeConfig
}

func (c *FakeConfig) Files() config.FileConfig {
	return &fileConfig{c}
}

func (fc *fileConfig) Templates() io.File {
	return fc.cfg.TemplateBase
}

func (fc *fileConfig) Content() io.File {
	return fc.cfg.InputBase
}

type siteConfig struct {
	cfg  *FakeConfig
	lang languages.Language
}

func (c *FakeConfig) Site(lang languages.Language) config.SiteConfig {
	return &siteConfig{c, lang}
}

func (sc *siteConfig) WebRoot() string {
	v, ok := sc.cfg.WebRoots[sc.lang.Code()]
	if !ok {
		return sc.cfg.WebRoots[""]
	}
	return v
}

func (sc *siteConfig) Name() string {
	v, ok := sc.cfg.SiteNames[sc.lang.Code()]
	if !ok {
		return sc.cfg.SiteNames[""]
	}
	return v
}

func (sc *siteConfig) Uri() string {
	v, ok := sc.cfg.SiteURIs[sc.lang.Code()]
	if !ok {
		return sc.cfg.SiteURIs[""]
	}
	return v
}

func (sc *siteConfig) Language() languages.Language {
	return sc.lang
}

type authorConfig struct {
	cfg *FakeConfig
}

func (c *FakeConfig) Author() config.AuthorConfig {
	return &authorConfig{c}
}

func (ac *authorConfig) Name() string {
	return ac.cfg.AuthorName
}

func (ac *authorConfig) Uri() string {
	return ac.cfg.AuthorURI
}

type generatorConfig struct {
	cfg *FakeConfig
}

func (c *FakeConfig) Generator() config.GeneratorConfig {
	return &generatorConfig{c}
}

func (gc *generatorConfig) Output() io.File {
	return gc.cfg.OutputBase
}

func (gc *generatorConfig) HideUntranslated() bool {
	return gc.cfg.HideUntranslated
}

func (gc *generatorConfig) PublishUntil() time.Time {
	return gc.cfg.PublishUntil
}

func (c *FakeConfig) Mailers() []config.MailerConfig {
	return []config.MailerConfig{}
}
