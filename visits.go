package beenthere

import (
	r "github.com/dancannon/gorethink"
)

type VisitClient struct {
	session *r.Session
}

type Visit struct {
	City  string `json:"city" xml:"city"`
	State string `json:"state" xml:"state"`
}

func NewVisitClient(sess *r.Session) *VisitClient {
	return &VisitClient{sess}
}

func NewVisit() *Visit {
	return &Visit{}
}

func (c *VisitClient) Validate(visit *Visit) error {
	return nil
}

func (c *VisitClient) GetUserVisits(userId string) ([]Visit, error) {
	return nil, nil
}

func (c *VisitClient) AddUserVisit(userId string, visit *Visit) error {
	return nil
}
