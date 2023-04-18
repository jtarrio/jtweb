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
	client      *mlgo.Client
	languageIds map[string]int
}

func (m *Mailerlite) parseUtcTime(t string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", t)
}

func (m *Mailerlite) getUtcTimezone() (int, error) {
	if m.utcTz != -1 {
		return m.utcTz, nil
	}

	ctx := context.TODO()
	tzs, _, err := m.client.Timezone.List(ctx)
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

func (m *Mailerlite) getLanguageId(language string) (int, error) {
	if len(m.languageIds) == 0 {
		ctx := context.TODO()
		lids, _, err := m.client.Campaign.Languages(ctx)
		if err != nil {
			return -1, err
		}
		for _, lid := range lids.Data {
			m.languageIds[lid.Shortcode], err = strconv.Atoi(lid.Id)
			if err != nil {
				return -1, err
			}
		}
	}
	var id int
	ok := false
	for !ok && language != "" {
		id, ok = m.languageIds[language]
		if !ok {
			if language == "gl" {
				language = "pt"
			} else {
				language = ""
			}
		}
	}
	if !ok {
		return -1, fmt.Errorf("language not found")
	}
	return id, nil
}

func ConnectMailerliteV2(apikey string, senderName string, senderEmail string, group string, dryRun bool) (*Mailerlite, error) {
	return &Mailerlite{
		apikey:      apikey,
		senderName:  senderName,
		senderEmail: senderEmail,
		group:       group,
		dryRun:      dryRun,
		utcTz:       -1,
		client:      mlgo.NewClient(apikey),
		languageIds: make(map[string]int),
	}, nil
}

func (m *Mailerlite) GetScheduledEmailDates() ([]email.ScheduledEmail, error) {
	ctx := context.TODO()
	campaigns, _, err := m.client.Campaign.List(ctx, &mlgo.ListCampaignOptions{
		Filters: &[]mlgo.Filter{
			{Name: "status", Value: "ready"},
			{Name: "type", Value: "regular"},
		},
		Page:  1,
		Limit: 100,
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
	html, err := inliner.Inline(email.Html)
	if err != nil {
		return "", err
	}

	language, err := m.getLanguageId(email.Language)
	if err != nil {
		return "", err
	}

	request := &mlgo.CreateCampaign{
		Name:       email.Name,
		Type:       mlgo.CampaignTypeRegular,
		LanguageID: language,
		Groups:     []string{m.group},
		Emails: []mlgo.Emails{
			{FromName: m.senderName,
				From:      m.senderEmail,
				Subject:   email.Subject,
				Content:   html,
				PlainText: email.Plaintext,
			},
		},
	}

	if m.dryRun {
		fakeId++
		// fmt.Println(*request)
		return fmt.Sprint(fakeId), nil
	}

	ctx := context.TODO()
	campaign, _, err := m.client.Campaign.Create(ctx, request)
	if err != nil {
		return "", err
	}

	id := campaign.Data.ID
	return id, nil
}

func (m *Mailerlite) sendOrSchedule(id string, date *time.Time) error {
	tzid, err := m.getUtcTimezone()
	if err != nil {
		return err
	}
	ctx := context.TODO()
	var schedule *mlgo.ScheduleCampaign
	if date != nil {
		schedule = &mlgo.ScheduleCampaign{
			Delivery: mlgo.CampaignScheduleTypeScheduled,
			Schedule: &mlgo.Schedule{
				Date:     date.UTC().Format("2006-01-02"),
				Hours:    fmt.Sprintf("%02d", date.UTC().Hour()),
				Minutes:  fmt.Sprintf("%02d", date.UTC().Minute()),
				Timezone: tzid,
			},
		}
	} else {
		schedule = &mlgo.ScheduleCampaign{
			Delivery: mlgo.CampaignScheduleTypeInstant,
		}
	}
	if m.dryRun {
		if schedule.Delivery == mlgo.CampaignScheduleTypeScheduled {
			fmt.Println(*schedule.Schedule)
		}
		return nil
	} else {
		_, _, err = m.client.Campaign.Schedule(ctx, id, schedule)
		return err
	}
}

func (m *Mailerlite) Schedule(id string, date time.Time) error {
	return m.sendOrSchedule(id, &date)
}

func (m *Mailerlite) Send(id string) error {
	return m.sendOrSchedule(id, nil)
}
