package cities

import (
	r "github.com/dancannon/gorethink"
	"github.com/dancannon/gorethink/types"
)

type Client struct {
	config  Config
	session *r.Session
}

type City struct {
	ID       string      `json:"id" xml:"id" gorethink:"id"`
	Name     string      `json:"name" xml:"name" gorethink:"name"`
	State    string      `json:"state" xml:"state" gorethink:"state"`
	Location types.Point `json:location xml:"location" gorethink:"location"`
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

func (c *Client) GetCities(state string) ([]string, error) {
	result, err := r.DB(c.config.DB).Table(c.config.Table).GetAllByIndex("state", state).Run(c.session)
	_, _ = result, err
	return nil, nil
}

func (c *Client) StateName(state string) string {
	return states[state]
}
