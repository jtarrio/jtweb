package site

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v2"
)

// Config contains the parameters of the current site.
type Config struct {
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

// ParseConfig reads the configuration from a file.
func ParseConfig(b []byte) (*Config, error) {
	cfg := &Config{}
	err := yaml.UnmarshalStrict(b, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// Normalize checks that the configuration is valid and fills any missing optional values.
func (c *Config) Normalize() error {
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

func (c *Config) GetWebRoot(lang string) string {
	uri, ok := c.WebRootLanguages[lang]
	if ok {
		return uri
	}
	return c.WebRoot
}

func (c *Config) GetSiteName(lang string) string {
	uri, ok := c.SiteNameLanguages[lang]
	if ok {
		return uri
	}
	return c.SiteName
}

func (c *Config) GetSiteURI(lang string) string {
	uri, ok := c.SiteURILanguages[lang]
	if ok {
		return uri
	}
	return c.SiteURI
}
