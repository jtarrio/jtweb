package email

import "time"

type ScheduledEmail struct {
	Id   int
	Name string
	When time.Time
}

type Email struct {
	Name      string
	Group     int
	Language  string
	Subject   string
	Plaintext string
	Html      string
}

type Mailer interface {
	GetScheduledEmailDates() ([]ScheduledEmail, error)
	DraftEmail(email Email) (int, error)
	Send(id int) error
	Schedule(id int, date time.Time) error
}
