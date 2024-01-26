package generator

import (
	"log"
	"strings"

	"jacobo.tarrio.org/jtweb/email"
	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/page"
	"jacobo.tarrio.org/jtweb/site"
)

// EmailGenerator extracts pages and converts them into emails.
type EmailGenerator struct {
	contents    *site.Contents
	language    languages.Language
	engine      email.Engine
	filter      func(*page.Page) bool
	makeName    func(*page.Page) string
	makeSubject func(*page.Page) string
	err         error
}

type EmailGeneratorOption func(*EmailGenerator)

// Creates a new EmailGenerator that will generate emails for the pages in the given site and language.
func NewEmailGenerator(c *site.Contents, language languages.Language, engine email.Engine) *EmailGenerator {
	return &EmailGenerator{c, language, engine, defaultFilter, defaultMakeName, defaultMakeSubject, nil}
}

// Adds options to the EmailGenerator.
func (g *EmailGenerator) WithOptions(options ...EmailGeneratorOption) *EmailGenerator {
	for _, option := range options {
		option(g)
	}
	return g
}

// Skips pages that already have a corresponding scheduled email.
func NotScheduled() EmailGeneratorOption {
	return func(g *EmailGenerator) {
		prev := g.filter
		campaigns, err := g.engine.ScheduledCampaigns()
		if err != nil {
			g.err = err
			return
		}
		seen := map[string]bool{}
		for _, campaign := range campaigns {
			seen[campaign.Name] = true
		}
		g.filter = func(p *page.Page) bool {
			_, ok := seen[g.makeName(p)]
			return !ok && prev(p)
		}
	}
}

// Adds a prefix to the email names.
func NamePrefix(prefix string) EmailGeneratorOption {
	return func(g *EmailGenerator) {
		prev := g.makeName
		g.makeName = func(p *page.Page) string {
			return prefix + " " + prev(p)
		}
	}
}

// Adds a prefix to the email subjects.
func SubjectPrefix(prefix string) EmailGeneratorOption {
	return func(g *EmailGenerator) {
		prev := g.makeSubject
		g.makeSubject = func(p *page.Page) string {
			if p.Header.Episode == "" {
				return prefix + ": " + prev(p)
			} else {
				return prefix + " " + prev(p)
			}
		}
	}
}

func (g *EmailGenerator) SendMails() error {
	if g.err != nil {
		return g.err
	}

	toc, ok := g.contents.Toc[g.language]
	if !ok {
		return nil
	}

	var emails []*email.Email
	for _, name := range toc.All {
		page := g.contents.Pages[name]
		if !g.filter(page) {
			continue
		}
		var e email.Email
		e.Name = g.makeName(page)
		e.Subject = g.makeSubject(page)
		e.Language = page.Header.Language
		e.Date = page.Header.PublishDate
		{
			sb := strings.Builder{}
			err := g.contents.OutputAsEmail(&sb, page)
			if err != nil {
				return err
			}
			e.Html = sb.String()
		}
		{
			sb := strings.Builder{}
			err := g.contents.OutputAsPlainEmail(&sb, page)
			if err != nil {
				return err
			}
			e.Plaintext = sb.String()
		}
		emails = append(emails, &e)
	}

	for _, email := range emails {
		campaign, err := g.engine.CreateCampaign(email)
		if err != nil {
			return err
		}
		err = campaign.Schedule()
		if err != nil {
			return err
		}
		log.Printf("Scheduled [%s] as id %s for %s", email.Name, campaign.Id(), email.Date.String())
	}

	return nil
}

func defaultFilter(*page.Page) bool {
	return true
}

func defaultMakeName(p *page.Page) string {
	return string(p.Name)
}

func defaultMakeSubject(p *page.Page) string {
	if p.Header.Episode == "" {
		return p.Header.Title
	} else {
		return p.Header.Episode + ": " + p.Header.Title
	}
}
