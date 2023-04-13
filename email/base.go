package email

import "time"

type ScheduledEmail struct {
	Id   string
	Name string
	When time.Time
}

type Email struct {
	Name      string
	Language  string
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
