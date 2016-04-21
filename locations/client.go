package locations

import (
	"errors"
	"fmt"
	"strings"

	r "github.com/dancannon/gorethink"
	"github.com/dancannon/gorethink/types"
	"github.com/nstogner/beenthere-ws/visits"
)

var (
	ErrAlreadyExists = errors.New("city already exists")
	ErrNoSuchState   = errors.New("no such state")
)

// City is a db structure.
type City struct {
	ID       string      `json:"id" xml:"id" gorethink:"id"`
	Name     string      `json:"name" xml:"name" gorethink:"name"`
	State    string      `json:"state" xml:"state" gorethink:"state"`
	Location types.Point `json:"location,omitempty" xml:"location,omitempty" gorethink:"location,omitempty"`
	Verified bool        `json:"-" xml:"-" gorethink:"verified"`
}

// Client acts as an api to retreiving city info from a db.
type Client struct {
	config  Config
	session *r.Session
}

// Config is used to create a new Client instance via NewClient(...).
type Config struct {
	Table string
}

// CityFromVisit returns a new City entity from a given Visit entity.
func CityFromVisit(v *visits.Visit) *City {
	id := v.City + "," + v.State
	return &City{
		ID:    id,
		Name:  v.City,
		State: v.State,
	}
}

// NewClient returns a instance of Client.
func NewClient(conf Config, sess *r.Session) *Client {
	return &Client{
		config:  conf,
		session: sess,
	}
}

// ValidateCity inspects the given City entity and returns a non-nil for an
// invalid entity. NOTE: This currently only verifies the State.
func (c *Client) ValidateCity(city *City) error {
	if c.StateName(city.State) == "" {
		return ErrNoSuchState
	}
	return nil
}

// Insert a new city into the database if the given city does not already
// exist.
func (c *Client) AddCity(city *City) error {
	_, err := r.Table(c.config.Table).Insert(city).RunWrite(c.session)
	if r.IsConflictErr(err) {
		return ErrAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("unable to add visit: %s", err.Error())
	}
	return nil
}

// GetCityNames returns a list of city names which are in the database and
// have the given state associated with them.
func (c *Client) GetCityNames(state string) ([]string, error) {
	result, err := r.Table(c.config.Table).GetAllByIndex("state", strings.ToUpper(state)).Field("city").Run(c.session)
	if err != nil {
		return nil, fmt.Errorf("unable to get cities: %s", err.Error())
	}
	cities := make([]string, 0)
	var ct string
	for result.Next(&ct) {
		cities = append(cities, ct)
	}
	return cities, nil
}

// StateName returns the name of a US state if it exists in a hardcoded map
// of in-memory states. The only argument is a 2-letter state abbreviation.
// If the state does not exist, an empty string is returned.
func (c *Client) StateName(state string) string {
	return states[strings.ToUpper(state)]
}
