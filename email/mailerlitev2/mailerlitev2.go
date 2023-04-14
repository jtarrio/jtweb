package mailerlitev2

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aymerick/douceur/inliner"
	mlgo "github.com/mailerlite/mailerlite-go"
	email "jacobo.tarrio.org/jtweb/email"
)

type Mailerlite struct {
	apikey      string
	senderName  string
	senderEmail string
	group       string
	dryRun      bool
	utcTz       int
}

func mapLanguage(language string) string {
	switch language {
	case "en":
	case "es":
		return language
	case "gl":
		return "pt"
	default:
		return "en"
	}
	return "en"
}

func getClient(m *Mailerlite) *mlgo.Client {
	return mlgo.NewClient(m.apikey)
}

func (m *Mailerlite) parseUtcTime(t string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", t)
}

func (m *Mailerlite) getUtcTimezone(client *mlgo.Client) (int, error) {
	if m.utcTz != -1 {
		return m.utcTz, nil
	}

	ctx := context.TODO()
	tzs, _, err := client.Timezone.List(ctx)
	if err != nil {
		return -1, err
	}

	for _, tz := range tzs.Data {
		if tz.Offset != 0 {
			continue
		}
		if strings.Contains(tz.Name, "Universal") ||
			strings.Contains(tz.Name, "UTC") ||
			strings.Contains(tz.Name, "Greenwich") ||
			strings.Contains(tz.Name, "GMT") {
			m.utcTz, err = strconv.Atoi(tz.Id)
			return m.utcTz, err
		}
	}

	return -1, fmt.Errorf("UTC timezone not found")
}

func ConnectMailerliteV2(apikey string, senderName string, senderEmail string, group string, dryRun bool) (*Mailerlite, error) {
	return &Mailerlite{apikey: apikey, senderName: senderName, senderEmail: senderEmail, group: group, dryRun: dryRun, utcTz: -1}, nil
}

func (m *Mailerlite) GetScheduledEmailDates() ([]email.ScheduledEmail, error) {
	client := getClient(m)
	ctx := context.TODO()
	campaigns, _, err := client.Campaign.List(ctx, &mlgo.ListCampaignOptions{
		Filters: &[]mlgo.Filter{
			{Name: "status", Value: "ready"},
			{Name: "type", Value: "regular"},
		},
		Page:  1,
		Limit: 1000,
	})
	if err != nil {
		return nil, err
	}

	var ret []email.ScheduledEmail
	for _, c := range campaigns.Data {
		dt, err := m.parseUtcTime(c.ScheduledFor)
		if err != nil {
			return nil, err
		}
		ret = append(ret, email.ScheduledEmail{Id: c.ID, Name: c.Name, When: dt})
	}

	return ret, nil
}

var fakeId int

func (m *Mailerlite) DraftEmail(email email.Email) (string, error) {
	if m.dryRun {
		fakeId++
		return fmt.Sprint(fakeId), nil
	}

	html, err := inliner.Inline(email.Html)
	if err != nil {
		return "", err
	}

	client := getClient(m)
	ctx := context.TODO()
	campaign, _, err := client.Campaign.Create(ctx, &mlgo.CreateCampaign{
		Name:   email.Name,
		Type:   mlgo.CampaignTypeRegular,
		Groups: []string{m.group},
		Emails: []mlgo.Emails{
			{FromName: m.senderName,
				From:    m.senderEmail,
				Subject: email.Subject,
				Content: html,
			},
		},
	})
	if err != nil {
		return "", err
	}

	id := campaign.Data.ID
	return id, nil
}

func (m *Mailerlite) sendOrSchedule(client *mlgo.Client, id string, date *time.Time) error {
	if m.dryRun {
		return nil
	}

	tzid, err := m.getUtcTimezone(client)
	if err != nil {
		return err
	}
	ctx := context.TODO()
	var schedule *mlgo.ScheduleCampaign
	if date != nil {
		schedule = &mlgo.ScheduleCampaign{
			Delivery: mlgo.CampaignScheduleTypeScheduled,
			Schedule: &mlgo.Schedule{
				Date:       date.UTC().Format("2006-01-02"),
				Hours:      fmt.Sprintf("%02d", date.Hour()),
				Minutes:    fmt.Sprintf("%02d", date.Minute()),
				TimezoneID: tzid,
			},
		}
	} else {
		schedule = &mlgo.ScheduleCampaign{
			Delivery: mlgo.CampaignScheduleTypeInstant,
		}
	}
	_, _, err = client.Campaign.Schedule(ctx, id, schedule)
	return err
}

func (m *Mailerlite) Schedule(id string, date time.Time) error {
	client := getClient(m)
	return m.sendOrSchedule(client, id, &date)
}

func (m *Mailerlite) Send(id string) error {
	client := getClient(m)
	return m.sendOrSchedule(client, id, nil)
}
