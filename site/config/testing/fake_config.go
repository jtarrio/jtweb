package testing

import (
	"time"

	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/site/io"
	"jacobo.tarrio.org/jtweb/site/io/testing"
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

func (c *FakeConfig) GetTemplateBase() io.File {
	return c.TemplateBase
}

func (c *FakeConfig) GetInputBase() io.File {
	return c.InputBase
}

func (c *FakeConfig) GetOutputBase() io.File {
	return c.OutputBase
}

func (c *FakeConfig) GetWebRoot(lang languages.Language) string {
	v, ok := c.WebRoots[lang.Code()]
	if !ok {
		return c.WebRoots[""]
	}
	return v
}

func (c *FakeConfig) GetSiteName(lang languages.Language) string {
	v, ok := c.SiteNames[lang.Code()]
	if !ok {
		return c.SiteNames[""]
	}
	return v
}

func (c *FakeConfig) GetSiteURI(lang languages.Language) string {
	v, ok := c.SiteURIs[lang.Code()]
	if !ok {
		return c.SiteURIs[""]
	}
	return v
}

func (c *FakeConfig) GetAuthorName() string {
	return c.AuthorName
}

func (c *FakeConfig) GetAuthorURI() string {
	return c.AuthorURI
}

func (c *FakeConfig) GetHideUntranslated() bool {
	return c.HideUntranslated
}

func (c *FakeConfig) GetPublishUntil() time.Time {
	return c.PublishUntil
}
