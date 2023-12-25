package lib

import (
	"jacobo.tarrio.org/jtweb/config"
	"jacobo.tarrio.org/jtweb/email/generator"
	"jacobo.tarrio.org/jtweb/site"
)

func OpEmail(mailer config.MailerConfig) OpFn {
	return func(rawContent *site.RawContents) error {
		notBefore := getTimeOrDefault(rawContent.Config.DateFilters().Mail().NotBefore(), rawContent.Config.DateFilters().Now())
		notAfter := rawContent.Config.DateFilters().Mail().NotAfter()
		content, err := rawContent.Index(notBefore, notAfter)
		if err != nil {
			return err
		}
		return generator.NewEmailGenerator(content, mailer.Language(), mailer.Engine()).
			WithOptions(
				generator.NotScheduled(),
				generator.NamePrefix(mailer.SubjectPrefix()),
				generator.SubjectPrefix(mailer.SubjectPrefix()),
			).SendMails()
	}
}
