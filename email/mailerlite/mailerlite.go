package mailerlite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	email "jacobo.tarrio.org/jtweb/email"
)

const baseUri = "https://api.mailerlite.com/api/"

type Mailerlite struct {
	apikey string
	group  int
	dryRun bool
	utcTz  int
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

func query[R any](m *Mailerlite, path string, method string, response *R) error {
	var dummy *string = nil

	return queryPayload(m, path, method, dummy, response)
}

func queryPayload[Q any, R any](m *Mailerlite, path string, method string, request *Q, response *R) error {
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
	body, err := ioutil.ReadAll(res.Body)
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

func (m *Mailerlite) parseUtcTime(t string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", t)
}

func (m *Mailerlite) getUtcTimezone() (int, error) {
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

func ConnectMailerlite(apikey string, group int, dryRun bool) (*Mailerlite, error) {
	return &Mailerlite{apikey: apikey, group: group, dryRun: dryRun, utcTz: -1}, nil
}

func (m *Mailerlite) GetScheduledEmailDates() ([]email.ScheduledEmail, error) {
	type campaign struct {
		Id       int    `json:"id"`
		DateSend string `json:"date_send"`
		Name     string `json:"name"`
	}

	var campaigns []campaign
	err := query(m, "v2/campaigns/outbox", "GET", &campaigns)
	if err != nil {
		return nil, err
	}

	var ret []email.ScheduledEmail
	for _, c := range campaigns {
		dt, err := m.parseUtcTime(c.DateSend)
		if err != nil {
			return nil, err
		}
		ret = append(ret, email.ScheduledEmail{Id: fmt.Sprint(c.Id), Name: c.Name, When: dt})
	}

	return ret, nil
}

var fakeId int

func (m *Mailerlite) DraftEmail(email email.Email) (string, error) {
	if m.dryRun {
		fakeId++
		return fmt.Sprint(fakeId), nil
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

		req := campaignReq{Type: "regular", Name: email.Name, Groups: []int{m.group}, Subject: email.Subject, Language: mapLanguage(email.Language)}
		var resp campaignResp
		err := queryPayload(m, "v2/campaigns", "POST", &req, &resp)
		if err != nil {
			return "", err
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

		req := contentReq{Plain: email.Plaintext, Html: email.Html, AutoInline: true}
		var resp contentResp
		err := queryPayload(m, fmt.Sprintf("v2/campaigns/%d/content", id), "PUT", &req, &resp)
		if err != nil {
			return "", err
		}
		if !resp.Success {
			return "", fmt.Errorf("Mailerlite returned a non-success state for id %d", id)
		}
	}

	return fmt.Sprint(id), nil
}

func (m *Mailerlite) sendOrSchedule(id int, date *time.Time) error {
	if m.dryRun {
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

	tzid, err := m.getUtcTimezone()
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
	return queryPayload(m, fmt.Sprintf("v2/campaigns/%d/actions/send", id), "POST", &req, &resp)
}

func (m *Mailerlite) Schedule(id string, date time.Time) error {
	idint, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	return m.sendOrSchedule(idint, &date)
}

func (m *Mailerlite) Send(id string) error {
	idint, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	return m.sendOrSchedule(idint, nil)
}
