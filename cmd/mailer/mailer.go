package main

import (
	"flag"

	"jacobo.tarrio.org/jtweb/config/fromflags"
	"jacobo.tarrio.org/jtweb/email/generator"
	"jacobo.tarrio.org/jtweb/site"
)

func main() {
	flag.Parse()

	cfg, err := fromflags.GetConfig()
	if err != nil {
		panic(err)
	}
	content, err := site.Read(cfg)
	if err != nil {
		panic(err)
	}

	for _, mailer := range cfg.Mailers() {
		err := generator.NewEmailGenerator(content, mailer.Language(), mailer.Engine()).
			WithOptions(
				generator.NotBefore(mailer.SendAfter()),
				generator.NotScheduled(),
				generator.NamePrefix(mailer.SubjectPrefix()),
				generator.SubjectPrefix(mailer.SubjectPrefix()),
			).SendMails()
		if err != nil {
			panic(err)
		}
	}
}
