package email

import "time"

type ScheduledEmail struct {
	Id   int
	Name string
	When time.Time
}

type Mailer interface {
	GetScheduledEmailDates() ([]ScheduledEmail, error)
	DraftEmail(name string, group int, subject string, plaintext string, html string) (int, error)
	Send(id int) error
	Schedule(id int, date time.Time) error
}
