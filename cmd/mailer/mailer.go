package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"jacobo.tarrio.org/jtweb/config/fromflags"
	"jacobo.tarrio.org/jtweb/email"
	"jacobo.tarrio.org/jtweb/email/generator"
	"jacobo.tarrio.org/jtweb/email/mailerlite"
	"jacobo.tarrio.org/jtweb/email/mailerlitev2"
	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/site"
)

var flagLanguage = flag.String("language", "", "The language to send mail for.")
var flagSendAfter = flag.String("send_after", "", "Schedule posts newer than the given date/time.")
var flagUseV2 = flag.Bool("use_v2", false, "Use Mailerlite V2")
var flagSenderName = flag.String("sender_name", "", "The sender's name.")
var flagSenderEmail = flag.String("sender_email", "", "The sender's email address.")
var flagApikey = flag.String("apikey", "", "The API key for the mailer.")
var flagGroup = flag.Int("group", -1, "The mail group to send to.")
var flagSubjectPrefix = flag.String("subject_prefix", "", "The prefix to use in subject lines.")
var flagSendFirstEmail = flag.Bool("send_first_email", false, "Instead of scheduling all emails, sends the first one.")
var flagLeaveAsDraft = flag.Bool("leave_as_draft", false, "Prepare the emails but don't schedule or send them.")
var flagDryRun = flag.Bool("dry_run", false, "Does not send or schedule emails.")

func getLanguage() (languages.Language, error) {
	if *flagLanguage == "" {
		return nil, fmt.Errorf("flag --language was not specified")
	}
	return languages.FindByCode(*flagLanguage)
}

func getEngine() (email.Engine, error) {
	if !*flagUseV2 {
		return mailerlite.ConnectMailerlite(*flagApikey, *flagGroup, *flagDryRun)
	}

	if *flagSenderName == "" || *flagSenderEmail == "" {
		return nil, fmt.Errorf("must specify a sender name and email")
	}
	return mailerlitev2.ConnectMailerliteV2(*flagApikey, *flagSenderName, *flagSenderEmail, fmt.Sprint(*flagGroup), *flagDryRun)
}

func getSendAfter() (time.Time, error) {
	if *flagSendAfter == "" {
		return time.Now(), nil
	}

	parsed, err := time.Parse(time.RFC3339, *flagSendAfter)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}

func sendMails(emails []*email.Email, engine email.Engine) error {
	if *flagSendFirstEmail {
		emails = []*email.Email{emails[len(emails)-1]}
	}

	for _, email := range emails {
		campaign, err := engine.CreateCampaign(email)
		if err != nil {
			return err
		}
		log.Printf("Created draft for [%s] as id %s", email.Name, campaign.Id())
		if *flagLeaveAsDraft {
			continue
		}
		if *flagSendFirstEmail {
			err = campaign.Send()
			if err != nil {
				return err
			}
			log.Printf("Sent id %s", campaign.Id())
		} else {
			err = campaign.Schedule()
			if err != nil {
				return err
			}
			log.Printf("Scheduled id %s for %s", campaign.Id(), email.Date.String())
		}
	}
	return nil
}

func main() {
	flag.Parse()

	language, err := getLanguage()
	if err != nil {
		panic(err)
	}

	sendAfter, err := getSendAfter()
	if err != nil {
		panic(err)
	}

	cfg, err := fromflags.GetConfig()
	if err != nil {
		panic(err)
	}
	content, err := site.Read(cfg)
	if err != nil {
		panic(err)
	}

	engine, err := getEngine()
	if err != nil {
		panic(err)
	}

	generator := generator.NewEmailGenerator(content, language).
		WithOptions(
			generator.NotBefore(sendAfter),
			generator.NotScheduled(engine),
			generator.NamePrefix(*flagSubjectPrefix),
			generator.SubjectPrefix(*flagSubjectPrefix),
		)
	emails, err := generator.CreateMails()
	if err != nil {
		panic(err)
	}

	if *flagDryRun {
		log.Print("Emails:")
		for _, email := range emails {
			log.Printf("%s\n"+"*** Language: %s\n"+"*** Date: %s\n"+"*** Subject: %s\n"+"*** HTML:\n%s\n"+"*** Text:\n%s",
				email.Name, email.Language.Code(), email.Date.String(), email.Subject, email.Html, email.Plaintext)
		}
	}

	if len(emails) == 0 {
		log.Print("No emails to be scheduled, exiting")
		return
	}

	err = sendMails(emails, engine)
	if err != nil {
		panic(err)
	}
}
