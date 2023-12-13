package yamlconfig

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v2"
	"jacobo.tarrio.org/jtweb/config"
	"jacobo.tarrio.org/jtweb/email"
	"jacobo.tarrio.org/jtweb/email/mailerlite"
	"jacobo.tarrio.org/jtweb/io"
	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/secrets"
)

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
		HideUntranslated bool `yaml:"hide_untranslated"`
		Disabled         bool
	}
	Mailers []struct {
		Name          string
		Language      string
		SubjectPrefix string `yaml:"subject_prefix"`
		Disabled      bool
		Mailerlite    *struct {
			ApikeySecret string `yaml:"apikey_secret"`
			Group        int
		}
	}
	DateFilters struct {
		Generate struct {
			NotAfter     *time.Time `yaml:"not_after"`
			NotAfterDays *uint      `yaml:"not_after_days"`
		}
		Mail struct {
			NotBefore    *time.Time `yaml:"not_before"`
			NotAfter     *time.Time `yaml:"not_after"`
			NotAfterDays *uint      `yaml:"not_after_days"`
		}
	} `yaml:"date_filters"`
	Debug struct {
		DryRun bool `yaml:"dry_run"`
	}
}

type configParser struct {
	source         []byte
	secretSupplier secrets.SecretSupplier
	options        []configParserOption
}

type configParserOption func(*yamlConfig)

func NewConfigParser(source []byte) *configParser {
	return &configParser{source: source, secretSupplier: secrets.NullSupplier()}
}

func (r *configParser) WithSecretSupplier(supplier secrets.SecretSupplier) *configParser {
	r.secretSupplier = secrets.CombineSuppliers(r.secretSupplier, supplier)
	return r
}

func (r *configParser) WithOptions(options ...configParserOption) *configParser {
	r.options = append(r.options, options...)
	return r
}

func OverrideWebroot(webroot string) configParserOption {
	return func(cfg *yamlConfig) {
		cfg.Site.Webroot = webroot
		for lang, langCfg := range cfg.Site.ByLanguage {
			langCfg.Webroot = webroot
			cfg.Site.ByLanguage[lang] = langCfg
		}
	}
}

func OverrideOutput(output string) configParserOption {
	return func(cfg *yamlConfig) {
		if cfg.Generator != nil {
			cfg.Generator.Output = output
		}
	}
}

func OverrideGenerateNotAfter(notAfter time.Time) configParserOption {
	return func(cfg *yamlConfig) {
		cfg.DateFilters.Generate.NotAfter = &notAfter
		cfg.DateFilters.Generate.NotAfterDays = nil
	}
}

func OverrideMailNotBefore(notBefore time.Time) configParserOption {
	return func(cfg *yamlConfig) {
		cfg.DateFilters.Mail.NotBefore = &notBefore
	}
}

func OverrideMailNotAfter(notAfter time.Time) configParserOption {
	return func(cfg *yamlConfig) {
		cfg.DateFilters.Mail.NotAfter = &notAfter
		cfg.DateFilters.Mail.NotAfterDays = nil
	}
}

func OverrideDryRun(dryRun bool) configParserOption {
	return func(cfg *yamlConfig) {
		cfg.Debug.DryRun = dryRun
	}
}

func (r *configParser) Parse() (config.Config, error) {
	var cfg = &yamlConfig{}
	err := yaml.UnmarshalStrict(r.source, cfg)
	if err != nil {
		return nil, err
	}

	for _, option := range r.options {
		option(cfg)
	}

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
			disabled:         cfg.Generator.Disabled,
		}
		if cfg.Debug.DryRun {
			out.generator.output = io.DryRunFile(out.generator.output)
		}
	}
	for _, mailer := range cfg.Mailers {
		if mailer.Name == "" {
			return nil, fmt.Errorf("the mailer's name has not been set")
		}
		outMailer := &mailerConfig{
			name:          mailer.Name,
			subjectPrefix: mailer.SubjectPrefix,
			disabled:      mailer.Disabled,
		}
		if mailer.Language == "" {
			return nil, fmt.Errorf("the mailer's language has not been set")
		}
		lang, err := languages.FindByCode(mailer.Language)
		if err != nil {
			return nil, err
		}
		outMailer.language = lang
		if mailer.Mailerlite != nil {
			apikey, err := r.secretSupplier.GetSecret(mailer.Mailerlite.ApikeySecret)
			if err != nil {
				return nil, err
			}
			engine, err := mailerlite.ConnectMailerlite(apikey, mailer.Mailerlite.Group, false)
			if err != nil {
				return nil, err
			}
			outMailer.engine = engine
		}
		if outMailer.engine == nil {
			return nil, fmt.Errorf("no email engine was defined")
		}
		if cfg.Debug.DryRun {
			outMailer.engine = email.DryRunEngine(outMailer.engine)
		}
		out.mailers = append(out.mailers, outMailer)
	}
	now := time.Now()
	out.dateFilters = dateFilterConfig{
		now: now,
		generate: dateFilter{
			notBefore: nil,
			notAfter:  parseRelDate(cfg.DateFilters.Generate.NotAfter, cfg.DateFilters.Generate.NotAfterDays, now),
		},
		mail: dateFilter{
			notBefore: cfg.DateFilters.Mail.NotBefore,
			notAfter:  parseRelDate(cfg.DateFilters.Mail.NotAfter, cfg.DateFilters.Mail.NotAfterDays, now),
		},
	}

	return out, nil
}

func parseRelDate(when *time.Time, days *uint, now time.Time) *time.Time {
	if when != nil {
		return when
	}
	if days != nil {
		hours := int(*days) * 24
		moment := now.Add(time.Duration(hours) * time.Hour)
		return &moment
	}
	return nil
}
