package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	r "github.com/dancannon/gorethink"

	"github.com/nstogner/beenthere-ws/handler"
	"github.com/nstogner/beenthere-ws/locations"
	"github.com/nstogner/beenthere-ws/visits"
)

// TestServer relies on rethinkdb being installed on the localhost. This can
// be changed by setting environment variables defined in config.go to point
// to another database.
func TestServer(t *testing.T) {
	// checkErr will fail the test for non-nil errors.
	checkErr := func(msg string, err error) {
		if err != nil {
			t.Fatalf("failure: %s: %s", msg, err.Error())
		}
	}

	conf := ConfigFromEnv()
	// This hardcoded db name is very important. It ensures that even if the
	// test is ran while configured to point to a production db thru env
	// variables, it wont harm the production data.
	conf.DBName = "testing"

	// Setup DB connection/clients.
	sess, err := r.Connect(r.ConnectOpts{
		Address:  fmt.Sprintf("%s:%s", conf.DBHost, conf.DBPort),
		Database: conf.DBName,
	})
	checkErr("connecting to db", err)
	vc := visits.NewClient(visits.Config{
		DB:    conf.DBName,
		Table: conf.VisitsTable,
	}, sess)
	lc := locations.NewClient(locations.Config{
		DB:    conf.DBName,
		Table: conf.CitiesTable,
	}, sess)

	// Setup http handler.
	hdlr := handler.New(handler.Config{
		Logger:       log,
		VisitsClient: vc,
		LocsClient:   lc,
	})
	server := httptest.NewServer(hdlr)

	// Setup test db/tables.
	_, err = r.TableCreate(conf.VisitsTable).RunWrite(sess)
	checkErr("creating table", err)
	_, err = r.TableCreate(conf.CitiesTable).RunWrite(sess)
	checkErr("creating table", err)
	_, err = r.Table(conf.VisitsTable).IndexCreate("user").RunWrite(sess)
	checkErr("creating table index", err)
	_, err = r.Table(conf.CitiesTable).IndexCreate("state").RunWrite(sess)
	checkErr("creating table index", err)
	defer func() {
		// Cleanup testing db.
		r.DBDrop(conf.DBName)
	}()

	// Run test cases.
	checkStatus := func(msg string, resp *http.Response, expect int) {
		if expect != resp.StatusCode {
			t.Fatalf("%s: expected http status code %v, got %v", msg, expect, resp.StatusCode)
		}
	}

	resp, err := http.Post(server.URL+"/users/testman/visits", "application/json", nil)
	checkErr("failed to make http request", err)
	checkStatus("POSTing an empty visit", resp, http.StatusBadRequest)
}
