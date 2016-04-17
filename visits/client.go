package visits

import (
	"fmt"
	"time"

	r "github.com/dancannon/gorethink"
)

type Client struct {
	config  Config
	session *r.Session
}

type Visit struct {
	ID        string    `json:"id" xml:"id" gorethink:"id,omitempty"`
	City      string    `json:"city,omitempty" xml:"city,omitempty" gorethink:"city"`
	State     string    `json:"state,omitempty" xml:"state,omitempty" gorethink:"state"`
	User      string    `json:"user,omitempty" xml:"user,omitempty" gorethink:"user"`
	Timestamp time.Time `json:"timestamp,omitempty" xml:"timestamp,omitempty" gorethink:"timestamp"`
}

type Config struct {
	DB    string
	Table string
}

func NewClient(conf Config, sess *r.Session) *Client {
	return &Client{
		config:  conf,
		session: sess,
	}
}

func NewVisit() *Visit {
	return &Visit{}
}

func (c *Client) Validate(visit *Visit) error {
	return nil
}

func (c *Client) GetVisits(userId string) ([]Visit, error) {
	result, err := r.DB(c.config.DB).Table(c.config.Table).GetAllByIndex("user", userId).Run(c.session)
	if err != nil {
		return nil, fmt.Errorf("unable to get visits: %s", err.Error())
	}
	visits := make([]Visit, 0)
	var v Visit
	for result.Next(&v) {
		visits = append(visits, v)
	}
	return visits, nil
}

func (c *Client) GetStates(userId string) ([]string, error) {
	result, err := r.DB(c.config.DB).Table(c.config.Table).GetAllByIndex("user", userId).Field("state").Distinct().Run(c.session)
	if err != nil {
		return nil, fmt.Errorf("unable to get visits: %s", err.Error())
	}
	states := make([]string, 0)
	var st string
	for result.Next(&st) {
		states = append(states, st)
	}
	return states, nil
}

func (c *Client) GetCities(userId string) ([]string, error) {
	result, err := r.DB(c.config.DB).Table(c.config.Table).GetAllByIndex("user", userId).Field("city").Distinct().Run(c.session)
	if err != nil {
		return nil, fmt.Errorf("unable to get visits: %s", err.Error())
	}
	cities := make([]string, 0)
	var ct string
	for result.Next(&ct) {
		cities = append(cities, ct)
	}
	return cities, nil
}

func (c *Client) Add(visit *Visit) error {
	visit.Timestamp = time.Now()
	result, err := r.DB(c.config.DB).Table(c.config.Table).Insert(visit).RunWrite(c.session)
	if err != nil {
		return fmt.Errorf("unable to add visit: %s", err.Error())
	}
	visit.ID = result.GeneratedKeys[0]
	return nil
}

func (c *Client) Delete(visitId string) error {
	_, err := r.DB(c.config.DB).Table(c.config.Table).Get(visitId).Delete().RunWrite(c.session)
	if err != nil {
		return fmt.Errorf("unable to delete visit: %s", err.Error())
	}
	return nil
}
