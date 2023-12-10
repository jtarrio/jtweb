package email

import (
	"fmt"
	"log"
)

type dryRunEngine struct {
	engine Engine
	lastId int64
}

func DryRunEngine(engine Engine) Engine {
	return &dryRunEngine{engine: engine, lastId: 1000}
}

func (e *dryRunEngine) ScheduledCampaigns() ([]ScheduledCampaign, error) {
	return e.engine.ScheduledCampaigns()
}

func (e *dryRunEngine) CreateCampaign(email *Email) (Campaign, error) {
	id := e.lastId + 1
	e.lastId = id
	log.Printf("[Dry run] Created campaign %d: %s", id, email.Name)
	return &nullCampaign{id: id, email: email}, nil
}

type nullCampaign struct {
	id    int64
	email *Email
}

func (c *nullCampaign) Id() string {
	return fmt.Sprint(c.id)
}

func (c *nullCampaign) Send() error {
	log.Printf("[Dry run] Sending campaign %d", c.id)
	return nil
}

func (c *nullCampaign) Schedule() error {
	log.Printf("[Dry run] Scheduling campaign %d for %s", c.id, c.email.Date.String())
	return nil
}
