package cities

import (
	"fmt"
	"strings"

	r "github.com/dancannon/gorethink"
	"github.com/dancannon/gorethink/types"
)

// Client acts as an api to retreiving city info from a db.
type Client struct {
	config  Config
	session *r.Session
}

// City is a db structure.
type City struct {
	ID       string      `json:"id" xml:"id" gorethink:"id"`
	Name     string      `json:"name" xml:"name" gorethink:"name"`
	State    string      `json:"state" xml:"state" gorethink:"state"`
	Location types.Point `json:location xml:"location" gorethink:"location"`
}

// Config is used to create a new Client instance via NewClient(...).
type Config struct {
	DB    string
	Table string
}

// NewClient returns a instance of Client.
func NewClient(conf Config, sess *r.Session) *Client {
	return &Client{
		config:  conf,
		session: sess,
	}
}

// GetCityNames returns a list of city names which are in the database and
// have the given state associated with them.
func (c *Client) GetCityNames(state string) ([]string, error) {
	result, err := r.DB(c.config.DB).Table(c.config.Table).GetAllByIndex("state", state).Run(c.session)
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
	return strings.ToUpper(states[state])
}
