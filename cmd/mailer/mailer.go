package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"jacobo.tarrio.org/jtweb/email"
	"jacobo.tarrio.org/jtweb/page"
	"jacobo.tarrio.org/jtweb/renderer/templates"
	"jacobo.tarrio.org/jtweb/site"
)

var flagLanguage = flag.String("language", "", "The language to send mail for.")
var flagSendAfter = flag.String("send_after", "", "Schedule posts newer than the given date/time.")
var flagApikey = flag.String("apikey", "", "The API key for the mailer.")
var flagGroup = flag.Int("group", -1, "The mail group to send to.")
var flagSubjectPrefix = flag.String("subject_prefix", "", "The prefix to use in subject lines.")
var flagSendFirstEmail = flag.Bool("send_first_email", false, "Instead of scheduling all emails, sends the first one.")
var flagLeaveAsDraft = flag.Bool("leave_as_draft", false, "Prepare the emails but don't schedule or send them.")
var flagDryRun = flag.Bool("dry_run", false, "Does not send or schedule emails.")

type emailData struct {
	subject   string
	date      time.Time
	plaintext string
	html      string
}

func gatherEmails(language string, sendAfter time.Time, subjectPrefix string, content *site.Contents, mailer email.Mailer) ([]emailData, error) {
	toc, ok := content.Toc[language]
	if !ok {
		return nil, fmt.Errorf("no table of contents for language: %s", language)
	}

	scheduled, err := mailer.GetScheduledEmailDates()
	if err != nil {
		return nil, err
	}

	seen := make(map[int64]string)
	for _, e := range scheduled {
		fmt.Printf("Found scheduled email [%s] for %s\n", e.Name, e.When.String())
		seen[e.When.Unix()] = e.Name
	}

	fmt.Printf("Finding pages scheduled after %s\n", sendAfter.String())
	var pages []*page.Page
	for _, name := range toc.All {
		page := content.Pages[name]
		if page.Header.PublishDate.After(sendAfter) {
			_, ok := seen[page.Header.PublishDate.Unix()]
			if ok {
				fmt.Printf("Page [%s] for %s already scheduled\n", page.Name, page.Header.PublishDate.String())
			} else {
				fmt.Printf("Schedule page [%s] for %s\n", page.Name, page.Header.PublishDate.String())
				pages = append(pages, page)
			}
		}
	}

	t := &templates.Templates{
		TemplatePath: content.TemplatePath,
		WebRoot:      content.GetWebRoot(language),
		Site:         templates.LinkData{Name: content.GetSiteName(language), URI: content.GetSiteURI(language)},
	}

	var emails []emailData
	for _, page := range pages {
		var email emailData
		if subjectPrefix == "" {
			email.subject = page.Header.Title
		} else if page.Header.Episode == "" {
			email.subject = subjectPrefix + ": " + page.Header.Title
		} else {
			email.subject = subjectPrefix + " " + page.Header.Episode + ": " + page.Header.Title
		}
		email.date = page.Header.PublishDate
		{
			sb := strings.Builder{}
			err := content.OutputEmail(&sb, t, page)
			if err != nil {
				return nil, err
			}
			email.html = sb.String()
		}
		{
			sb := strings.Builder{}
			err = content.OutputPlainEmail(&sb, t, page)
			if err != nil {
				return nil, err
			}
			email.plaintext = sb.String()
		}
		emails = append(emails, email)
	}

	return emails, nil
}

func draftEmail(m email.Mailer, group int, language string, data *emailData) (int, error) {
	name, _, _ := strings.Cut(data.subject, ":")
	id, err := m.DraftEmail(email.Email{Name: name, Group: group, Language: language, Subject: data.subject, Plaintext: data.plaintext, Html: data.html})
	if err != nil {
		return -1, err
	}
	fmt.Printf("Created draft for [%s] as id %d\n", name, id)
	return id, nil
}

func main() {
	flag.Parse()

	var language string
	if *flagLanguage != "" {
		language = *flagLanguage
	} else {
		panic(fmt.Errorf("--language was not specified"))
	}

	sendAfter := time.Now()
	if *flagSendAfter != "" {
		parsed, err := time.Parse(time.RFC3339, *flagSendAfter)
		if err != nil {
			panic(err)
		}
		sendAfter = parsed
	}

	cfg, err := site.FromFlags()
	if err != nil {
		panic(err)
	}
	content, err := cfg.Read()
	if err != nil {
		panic(err)
	}

	mailer, err := email.ConnectMailerlite(*flagApikey, *flagGroup, *flagDryRun)
	if err != nil {
		panic(err)
	}

	emails, err := gatherEmails(language, sendAfter, *flagSubjectPrefix, content, mailer)
	if err != nil {
		panic(err)
	}

	if len(emails) == 0 {
		fmt.Printf("No emails to be scheduled, exiting\n")
		return
	}

	if *flagSendFirstEmail {
		id, err := draftEmail(mailer, *flagGroup, language, &emails[len(emails)-1])
		if err != nil {
			panic(err)
		}
		if *flagLeaveAsDraft {
			return
		}
		err = mailer.Send(id)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Sent id %d\n", id)
		return
	}

	for _, email := range emails {
		id, err := draftEmail(mailer, *flagGroup, language, &email)
		if err != nil {
			panic(err)
		}
		if *flagLeaveAsDraft {
			continue
		}
		err = mailer.Schedule(id, email.date)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Scheduled id %d for %s\n", id, email.date.String())
	}
}
