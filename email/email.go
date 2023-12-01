// The email package has infrastructure to send out email updates when new
// posts are published.
//
// You start by defining an "email" for the post, with a human-readable name,
// language, subject line, and HTML and plain text contents.
// Then you create a "campaign" for that email.
// Finally, you can send that campaign immediately or schedule it campaign to go out on a particular day and time.

package email

import (
	"time"

	"jacobo.tarrio.org/jtweb/languages"
)

// Email contains the data for a particular email.
type Email struct {
	// A human-readable name for the email.
	Name string
	// The language the email is written in.
	Language languages.Language
	// The email's publish date
	Date time.Time
	// The subject line for the email.
	Subject string
	// The HTML content of the email.
	Html string
	// The plain text content of the email.
	Plaintext string
}

// ScheduledCampaign contains the data for a scheduled campaign.
type ScheduledCampaign struct {
	// The campaign's name.
	Name string
	// The time the campaign is scheduled for.
	When time.Time
}

// Engine provides methods to create new campaigns and list existing campaigns.
type Engine interface {
	// Returns a list of scheduled campaigns.
	ScheduledCampaigns() ([]ScheduledCampaign, error)
	// Creates a new campaign.
	CreateCampaign(email *Email) (Campaign, error)
}

// Campaign provides methods to send or schedule a campaign.
type Campaign interface {
	// Returns an identifier for the campaign.
	Id() string
	// Sends the campaign immediately.
	Send() error
	// Schedules the campaign for the email's publish date. It is undefined what happens if the date has already passed.
	Schedule() error
}
