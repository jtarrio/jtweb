package site

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// Config contains the parameters of the current site.
type Config struct {
	TemplatePath string `yaml:"template_path"`
	InputPath    string `yaml:"input_path"`
	OutputPath   string `yaml:"output_path"`
	WebRoot      string `yaml:"webroot"`
	SiteName     string `yaml:"site_name"`
	SiteURI      string `yaml:"site_uri"`
	AuthorName   string `yaml:"author_name"`
	AuthorURI    string `yaml:"author_uri"`
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
	return nil
}
