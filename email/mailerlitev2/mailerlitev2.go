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
	"jacobo.tarrio.org/jtweb/languages"
)

func ConnectMailerliteV2(apikey string, senderName string, senderEmail string, group string, dryRun bool) (*mailerlite, error) {
	return &mailerlite{
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

type mailerlite struct {
	apikey      string
	senderName  string
	senderEmail string
	group       string
	dryRun      bool
	utcTz       int
	client      *mlgo.Client
	languageIds map[string]int
}

type campaign struct {
	id   string
	when time.Time
	ml   *mailerlite
}

func (m *mailerlite) parseUtcTime(t string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", t)
}

func (m *mailerlite) getUtcTimezone() (int, error) {
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

func (m *mailerlite) getLanguageId(language languages.Language) (int, error) {
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
	code := language.Code()
	for !ok && code != "" {
		id, ok = m.languageIds[code]
		if !ok {
			if code == "gl" {
				code = "pt"
			} else {
				code = ""
			}
		}
	}
	if !ok {
		return -1, fmt.Errorf("language not found")
	}
	return id, nil
}

func (m *mailerlite) ScheduledCampaigns() ([]email.ScheduledCampaign, error) {
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

	var ret []email.ScheduledCampaign
	for _, c := range campaigns.Data {
		dt, err := m.parseUtcTime(c.ScheduledFor)
		if err != nil {
			return nil, err
		}
		ret = append(ret, email.ScheduledCampaign{Name: c.Name, When: dt})
	}

	return ret, nil
}

var fakeId int

func (m *mailerlite) CreateCampaign(email *email.Email) (email.Campaign, error) {
	html, err := inliner.Inline(email.Html)
	if err != nil {
		return nil, err
	}

	language, err := m.getLanguageId(email.Language)
	if err != nil {
		return nil, err
	}

	request := &mlgo.CreateCampaign{
		Name:       email.Name,
		Type:       mlgo.CampaignTypeRegular,
		LanguageID: language,
		Groups:     []string{m.group},
		Emails: []mlgo.Emails{{
			FromName:  m.senderName,
			From:      m.senderEmail,
			Subject:   email.Subject,
			Content:   html,
			PlainText: email.Plaintext,
		}},
	}

	if m.dryRun {
		fakeId++
		return &campaign{id: fmt.Sprint(fakeId), when: email.Date, ml: m}, nil
	}

	ctx := context.TODO()
	c, _, err := m.client.Campaign.Create(ctx, request)
	if err != nil {
		return nil, err
	}

	return &campaign{id: c.Data.ID, when: email.Date, ml: m}, nil
}

func (c *campaign) sendOrSchedule(date *time.Time) error {
	if c.ml.dryRun {
		return nil
	}

	tzid, err := c.ml.getUtcTimezone()
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
	_, _, err = c.ml.client.Campaign.Schedule(ctx, c.id, schedule)
	return err
}

func (c *campaign) Id() string {
	return c.id
}

func (c *campaign) Schedule() error {
	return c.sendOrSchedule(&c.when)
}

func (c *campaign) Send() error {
	return c.sendOrSchedule(nil)
}
