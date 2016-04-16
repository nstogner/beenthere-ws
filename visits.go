package main

import (
	"fmt"

	r "github.com/dancannon/gorethink"
)

type VisitClient struct {
	config  VisitConfig
	session *r.Session
}

type Visit struct {
	City  string `json:"city" xml:"city" gorethink:"city"`
	State string `json:"state" xml:"state" gorethink:"state"`
	User  string `json:"user" xml:"user" gorethink:"user"`
}

type VisitConfig struct {
	DB    string
	Table string
}

func NewVisitClient(conf VisitConfig, sess *r.Session) *VisitClient {
	return &VisitClient{
		config:  conf,
		session: sess,
	}
}

func NewVisit() *Visit {
	return &Visit{}
}

func (c *VisitClient) Validate(visit *Visit) error {
	return nil
}

func (c *VisitClient) GetFields(userId string, field string) ([]string, error) {
	result, err := r.DB(c.config.DB).Table(c.config.Table).GetAllByIndex("user", userId).Field(field).Run(c.session)
	if err != nil {
		return nil, fmt.Errorf("unable to get user visits: %s", err.Error())
	}
	fields := make([]string, 0)
	var f string
	for result.Next(&f) {
		fields = append(fields, f)
	}
	return fields, nil
}

func (c *VisitClient) Add(visit *Visit) error {
	_, err := r.DB(c.config.DB).Table(c.config.Table).Insert(visit).RunWrite(c.session)
	if err != nil {
		return fmt.Errorf("unable to add user visit: %s", err.Error())
	}
	return nil
}

func (c *VisitClient) Delete(visitId string) error {
	// TODO
	return nil
}
