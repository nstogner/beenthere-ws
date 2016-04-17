package visits

import (
	"fmt"
	"time"

	r "github.com/dancannon/gorethink"
)

// Client acts as an api to retreiving user visit info from a db.
type Client struct {
	config  Config
	session *r.Session
}

// Visit is a db structure for a single user visit to a specific city/state
// at a given time.
type Visit struct {
	ID        string    `json:"id" xml:"id" gorethink:"id,omitempty"`
	City      string    `json:"city,omitempty" xml:"city,omitempty" gorethink:"city"`
	State     string    `json:"state,omitempty" xml:"state,omitempty" gorethink:"state"`
	User      string    `json:"user,omitempty" xml:"user,omitempty" gorethink:"user"`
	Timestamp time.Time `json:"timestamp,omitempty" xml:"timestamp,omitempty" gorethink:"timestamp"`
}

// Config is used to create a new instance of Client via NewClient(...).
type Config struct {
	DB    string
	Table string
}

// NewClient returns a new instance of Client.
func NewClient(conf Config, sess *r.Session) *Client {
	return &Client{
		config:  conf,
		session: sess,
	}
}

// NewVisit returns a pointer to a new instance of Visit with Timestamp
// initialized to time.Now().
func NewVisit() *Visit {
	return &Visit{
		Timestamp: time.Now(),
	}
}

// Validate returns a non-nil error when it has been passed an invalid Visit
// entity.
func (c *Client) Validate(visit *Visit) error {
	return nil
}

// GetVisits gets a list of Visit entities from the database.
func (c *Client) GetVisits(userId string, start, limit int) ([]Visit, error) {
	result, err := r.DB(c.config.DB).Table(c.config.Table).GetAllByIndex("user", userId).Slice(start, start+limit).Run(c.session)
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

// GetStates gets a unique list of states visited by a given user from the
// database.
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

// GetCities gets a unique list of cities visited by a given user from the
// database.
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

// Add inserts a new Visit instance into the database.
func (c *Client) Add(visit *Visit) error {
	result, err := r.DB(c.config.DB).Table(c.config.Table).Insert(visit).RunWrite(c.session)
	if err != nil {
		return fmt.Errorf("unable to add visit: %s", err.Error())
	}
	visit.ID = result.GeneratedKeys[0]
	return nil
}

// Delete removes a Visit instance from the database given a unique visitId.
func (c *Client) Delete(visitId string) error {
	_, err := r.DB(c.config.DB).Table(c.config.Table).Get(visitId).Delete().RunWrite(c.session)
	if err != nil {
		return fmt.Errorf("unable to delete visit: %s", err.Error())
	}
	return nil
}
