package email

import (
	"time"

	"jacobo.tarrio.org/jtweb/languages"
)

type ScheduledEmail struct {
	Id   string
	Name string
	When time.Time
}

type Email struct {
	Name      string
	Language  languages.Language
	Subject   string
	Plaintext string
	Html      string
}

type Mailer interface {
	GetScheduledEmailDates() ([]ScheduledEmail, error)
	DraftEmail(email Email) (string, error)
	Send(id string) error
	Schedule(id string, date time.Time) error
}
