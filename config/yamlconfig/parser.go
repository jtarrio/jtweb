package yamlconfig

import (
	"bytes"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
	comments_engine "jacobo.tarrio.org/jtweb/comments/engine"
	"jacobo.tarrio.org/jtweb/comments/engine/mysql"
	"jacobo.tarrio.org/jtweb/comments/engine/sqlite3"
	email_notification "jacobo.tarrio.org/jtweb/comments/notification/email"
	comments_service "jacobo.tarrio.org/jtweb/comments/service"
	"jacobo.tarrio.org/jtweb/config"
	"jacobo.tarrio.org/jtweb/email"
	"jacobo.tarrio.org/jtweb/email/mailerlite"
	"jacobo.tarrio.org/jtweb/io"
	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/page"
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
		SkipOperation    bool `yaml:"skip_operation"`
	}
	Mailers []struct {
		Name          string
		Language      string
		SubjectPrefix string `yaml:"subject_prefix"`
		SkipOperation bool   `yaml:"skip_operation"`
		Mailerlite    *struct {
			ApikeySecret string `yaml:"apikey_secret"`
			Group        int
		}
	}
	Comments *struct {
		DefaultSetting      string `yaml:"default_setting"`
		PostAsDraft         bool   `yaml:"post_as_draft"`
		WidgetUri           string `yaml:"widget_uri"`
		AdminPasswordSecret string `yaml:"admin_password_secret"`
		SkipOperation       bool   `yaml:"skip_operation"`
		Sqlite3             *struct {
			ConnectionStringSecret string `yaml:"connection_string_secret"`
		}
		Mysql *struct {
			ConnectionStringSecret string `yaml:"connection_string_secret"`
		}
		Notify *struct {
			Email *struct {
				From           string
				To             string
				Host           string
				User           string
				PasswordSecret string `yaml:"password_secret"`
				Authentication string
				Encryption     string
			}
		}
	}
	DateFilters struct {
		Generate struct {
			NotAfter     *time.Time `yaml:"not_after"`
			NotAfterDays *int       `yaml:"not_after_days"`
		}
		Mail struct {
			NotBefore    *time.Time `yaml:"not_before"`
			NotAfter     *time.Time `yaml:"not_after"`
			NotAfterDays *int       `yaml:"not_after_days"`
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

func DisableComments() configParserOption {
	return func(cfg *yamlConfig) {
		cfg.Comments = nil
	}
}

func OverrideDryRun(dryRun bool) configParserOption {
	return func(cfg *yamlConfig) {
		cfg.Debug.DryRun = dryRun
	}
}

func (r *configParser) Parse() (config.Config, error) {
	var cfg = yamlConfig{}
	decoder := yaml.NewDecoder(bytes.NewReader(r.source))
	decoder.KnownFields(true)
	err := decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	for _, option := range r.options {
		option(&cfg)
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
			skipOperation:    cfg.Generator.SkipOperation,
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
			skipOperation: mailer.SkipOperation,
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
	if cfg.Comments != nil {
		var engine comments_engine.Engine = nil
		if cfg.Comments.Sqlite3 != nil {
			connString, err := r.secretSupplier.GetSecret(cfg.Comments.Sqlite3.ConnectionStringSecret)
			if err != nil {
				return nil, err
			}
			engine, err = sqlite3.NewSqlite3Engine(connString)
			if err != nil {
				return nil, err
			}
		} else if cfg.Comments.Mysql != nil {
			connString, err := r.secretSupplier.GetSecret(cfg.Comments.Mysql.ConnectionStringSecret)
			if err != nil {
				return nil, err
			}
			engine, err = mysql.NewMysqlEngine(connString)
			if err != nil {
				return nil, err
			}
		}
		defCfg, err := page.ParseCommentConfig(cfg.Comments.DefaultSetting)
		if err != nil {
			return nil, err
		}
		if defCfg == nil {
			defCfg = &page.CommentConfig{Enabled: false, Writable: false}
		}
		options := []comments_service.CommentsServiceOptions{
			comments_service.WithRenderer(comments_service.NewMarkdownRenderer()),
		}
		if cfg.Comments.PostAsDraft {
			options = append(options, comments_service.PostAsDraft())
		}
		adminPassword, err := r.secretSupplier.GetSecret(cfg.Comments.AdminPasswordSecret)
		if err != nil {
			return nil, err
		}
		if cfg.Comments.WidgetUri == "" {
			return nil, fmt.Errorf("no comments widget URI was defined")
		}
		if cfg.Comments.Notify != nil {
			if cfg.Comments.Notify.Email != nil {
				email := cfg.Comments.Notify.Email
				if email.From == "" {
					return nil, fmt.Errorf("no email 'from' address was specified")
				}
				if email.To == "" {
					return nil, fmt.Errorf("no email 'to' address was specified")
				}
				enc, err := email_notification.Encryption(email.Encryption)
				if err != nil {
					return nil, err
				}
				auth, err := email_notification.AuthType(email.Authentication)
				if err != nil {
					return nil, err
				}
				notifyOpts := []email_notification.NotificationEngineOption{
					email_notification.SetAuthType(auth),
					email_notification.SetEncryption(enc),
				}
				if email.User != "" {
					pass, err := r.secretSupplier.GetSecret(email.PasswordSecret)
					if err != nil {
						return nil, err
					}
					notifyOpts = append(notifyOpts, email_notification.SetAuth(email.User, pass))
				}
				if email.Host != "" {
					notifyOpts = append(notifyOpts, email_notification.SetHostPort(email.Host))
				}
				notify := email_notification.NewEmailNotificationEngine(
					appendToUri(cfg.Comments.WidgetUri, "admin.html"),
					email.From,
					email.To,
					notifyOpts...)
				options = append(options, comments_service.WithNotificationEngine(notify))
			}
		}
		out.comments = &commentsConfig{
			defaultConfig: defCfg,
			jsUri:         appendToUri(cfg.Comments.WidgetUri, "comments.js"),
			service:       comments_service.NewCommentsService(engine, options...),
			adminPassword: adminPassword,
			skipOperation: cfg.Comments.SkipOperation,
		}
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

func parseRelDate(when *time.Time, days *int, now time.Time) *time.Time {
	if when != nil {
		return when
	}
	if days != nil {
		hours := *days * 24
		moment := now.Add(time.Duration(hours) * time.Hour)
		return &moment
	}
	return nil
}

func appendToUri(uri string, path string) string {
	if uri == "" {
		return path
	}
	if uri[len(uri)-1] == '/' {
		return uri + path
	}
	return uri + "/" + path
}
