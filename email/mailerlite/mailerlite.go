package mailerlite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"jacobo.tarrio.org/jtweb/email"
	"jacobo.tarrio.org/jtweb/languages"
)

func ConnectMailerlite(apikey string, group int, dryRun bool) (*mailerlite, error) {
	return &mailerlite{apikey: apikey, group: group, dryRun: dryRun, utcTz: -1}, nil
}

const baseUri = "https://api.mailerlite.com/api/"

type mailerlite struct {
	apikey string
	group  int
	dryRun bool
	utcTz  int
}

type campaign struct {
	id   int
	when time.Time
	ml   *mailerlite
}

func mapLanguage(language languages.Language) string {
	switch language.Code() {
	case "en":
	case "es":
		return language.Code()
	case "gl":
		return "pt"
	default:
		return "en"
	}
	return "en"
}

func query[R any](m *mailerlite, path string, method string, response *R) error {
	var dummy *string = nil

	return queryPayload(m, path, method, dummy, response)
}

func queryPayload[Q any, R any](m *mailerlite, path string, method string, request *Q, response *R) error {
	url := baseUri + path

	var payload io.Reader
	if request == nil {
		payload = nil
	} else {
		j, err := json.Marshal(request)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(j)
	}

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-MailerLite-ApiKey", m.apikey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("request for %s returned status \"%s\": %s", req.URL, res.Status, string(body))
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	return nil
}

func (m *mailerlite) parseUtcTime(t string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", t)
}

func (m *mailerlite) getUtcTimezone() (int, error) {
	if m.utcTz >= 0 {
		return m.utcTz, nil
	}

	type tzdata struct {
		Id     int    `json:"id"`
		Offset int    `json:"time"`
		Name   string `json:"title"`
	}

	var tzs []tzdata
	err := query(m, "master/timezones", "GET", &tzs)
	if err != nil {
		return -1, err
	}

	for _, tz := range tzs {
		if tz.Offset != 0 {
			continue
		}
		if strings.Contains(tz.Name, "Universal") ||
			strings.Contains(tz.Name, "UTC") ||
			strings.Contains(tz.Name, "Greenwich") ||
			strings.Contains(tz.Name, "GMT") {
			m.utcTz = tz.Id
			return tz.Id, nil
		}
	}

	return -1, fmt.Errorf("UTC timezone not found")
}

func (m *mailerlite) Name() string {
	return "mailerlite"
}

func (m *mailerlite) ScheduledCampaigns() ([]email.ScheduledCampaign, error) {
	type campaign struct {
		DateSend string `json:"date_send"`
		Name     string `json:"name"`
	}

	var campaigns []campaign
	err := query(m, "v2/campaigns/outbox", "GET", &campaigns)
	if err != nil {
		return nil, err
	}

	var ret []email.ScheduledCampaign
	for _, c := range campaigns {
		dt, err := m.parseUtcTime(c.DateSend)
		if err != nil {
			return nil, err
		}
		ret = append(ret, email.ScheduledCampaign{
			Name: c.Name,
			When: dt,
		})
	}

	return ret, nil
}

var fakeId int

func (m *mailerlite) CreateCampaign(e *email.Email) (email.Campaign, error) {
	if m.dryRun {
		fakeId++
		return &campaign{id: fakeId, when: e.Date, ml: m}, nil
	}

	var id int

	{
		type campaignReq struct {
			Type     string `json:"type"`
			Name     string `json:"name"`
			Groups   []int  `json:"groups"`
			Subject  string `json:"subject"`
			Language string `json:"language"`
		}

		type campaignResp struct {
			Id int `json:"id"`
		}

		req := campaignReq{
			Type:     "regular",
			Name:     e.Name,
			Groups:   []int{m.group},
			Subject:  e.Subject,
			Language: mapLanguage(e.Language),
		}
		var resp campaignResp
		err := queryPayload(m, "v2/campaigns", "POST", &req, &resp)
		if err != nil {
			return nil, err
		}
		id = resp.Id
	}

	{
		type contentReq struct {
			Plain      string `json:"plain"`
			Html       string `json:"html"`
			AutoInline bool   `json:"auto_inline"`
		}

		type contentResp struct {
			Success bool `json:"success"`
		}

		req := contentReq{
			Plain:      e.Plaintext,
			Html:       e.Html,
			AutoInline: false,
		}
		var resp contentResp
		err := queryPayload(m, fmt.Sprintf("v2/campaigns/%d/content", id), "PUT", &req, &resp)
		if err != nil {
			return nil, err
		}
		if !resp.Success {
			return nil, fmt.Errorf("non-success state returned by Mailerlite for id %d", id)
		}
	}

	return &campaign{id: id, when: e.Date, ml: m}, nil
}

func (c *campaign) sendOrSchedule(date *time.Time) error {
	if c.ml.dryRun {
		return nil
	}

	type actionReq struct {
		Type      int    `json:"type"`
		Analytics int    `json:"analytics"`
		Date      string `json:"date,omitempty"`
		TzId      int    `json:"timezone_id,omitempty"`
	}

	type actionResp struct {
		Id int `json:"id"`
	}

	tzid, err := c.ml.getUtcTimezone()
	if err != nil {
		return err
	}

	req := actionReq{Analytics: 1}
	if date == nil {
		req.Type = 1
	} else {
		req.Type = 2
		req.Date = date.UTC().Format("2006-01-02 15:04")
		req.TzId = tzid
	}
	var resp actionResp
	return queryPayload(c.ml, fmt.Sprintf("v2/campaigns/%d/actions/send", c.id), "POST", &req, &resp)
}

func (c *campaign) Id() string {
	return fmt.Sprint(c.id)
}

func (c *campaign) Schedule() error {
	return c.sendOrSchedule(&c.when)
}

func (c *campaign) Send() error {
	return c.sendOrSchedule(nil)
}
