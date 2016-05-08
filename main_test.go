package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	r "github.com/dancannon/gorethink"

	"github.com/nstogner/beenthere-ws/handler"
	"github.com/nstogner/beenthere-ws/locations"
	"github.com/nstogner/beenthere-ws/visits"
)

// TestServer relies on rethinkdb being installed on the localhost. This can
// be changed by setting environment variables defined in config.go to point
// to another database. To install: https://www.rethinkdb.com/docs/install/
func TestServer(t *testing.T) {
	// checkErr will fail the test for non-nil errors.
	checkErr := func(msg string, err error) {
		if err != nil {
			t.Fatalf("failure: %s: %s", msg, err.Error())
		}
	}

	// This hardcoded db name is very important. It ensures that even if the
	// test is ran while configured to point to a production db thru env
	// variables, it wont harm the production data.
	os.Setenv("DB_NAME", "beenthere_testing")
	conf := ConfigFromEnv()
	// Really make sure it is the testing db...
	conf.DBName = "beenthere_testing"

	// Setup DB connection/clients.
	sess, err := r.Connect(r.ConnectOpts{
		Address:  fmt.Sprintf("%s:%s", conf.DBHost, conf.DBPort),
		Database: conf.DBName,
	})
	checkErr("connecting to db", err)
	vc := visits.NewClient(visits.Config{
		Table: conf.VisitsTable,
	}, sess)
	lc := locations.NewClient(locations.Config{
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
	// Drop the db in case the last test did not get the chance to cleanup...
	r.DBDrop(conf.DBName).RunWrite(sess)
	_, err = r.DBCreate(conf.DBName).RunWrite(sess)
	checkErr("creating db", err)
	_, err = r.TableCreate(conf.VisitsTable).RunWrite(sess)
	checkErr("creating table", err)
	_, err = r.TableCreate(conf.CitiesTable).RunWrite(sess)
	checkErr("creating table", err)
	_, err = r.Table(conf.VisitsTable).IndexCreate("user").RunWrite(sess)
	checkErr("creating table index", err)
	_, err = r.Table(conf.CitiesTable).IndexCreate("state").RunWrite(sess)
	checkErr("creating table index", err)
	_, err = r.Table(conf.VisitsTable).IndexWait("user").Run(sess)
	checkErr("waiting on table index", err)
	_, err = r.Table(conf.CitiesTable).IndexWait("state").Run(sess)
	checkErr("waiting on table index", err)
	_, err = r.Table(conf.CitiesTable).Insert(map[string]string{
		"id":    "Raleigh,NC",
		"state": "NC",
		"city":  "Raleigh",
	}).RunWrite(sess)
	checkErr("inserting city record", err)
	_, err = r.Table(conf.CitiesTable).Insert(map[string]string{
		"id":    "Charlotte,NC",
		"state": "NC",
		"city":  "Charlotte",
	}).RunWrite(sess)
	checkErr("inserting city record", err)
	defer func() {
		// Cleanup testing db.
		r.DBDrop(conf.DBName).RunWrite(sess)
	}()

	// Run test cases.
	checkStatus := func(msg string, resp *http.Response, expect int) {
		if expect != resp.StatusCode {
			t.Fatalf("%s: expected http status code %v, got %v", msg, expect, resp.StatusCode)
		}
	}

	// TEST CASES:

	// Start a streaming client.
	var scanner *bufio.Scanner
	var streamResp *http.Response
	go func() {
		var streamErr error
		streamResp, streamErr = http.Get(server.URL + "/stream/visits")
		checkErr("making http request", streamErr)
		scanner = bufio.NewScanner(streamResp.Body)
	}()

	// Add a user visit.
	resp, err := http.Post(
		server.URL+"/users/testman/visits",
		"application/json",
		// POST a lowercase state and verify later that it was converted to
		// uppercase.
		strings.NewReader(`{"city": "Raleigh", "state": "nc"}`),
	)
	checkErr("making http request", err)
	checkStatus("POSTing a valid visit", resp, http.StatusOK)
	resp.Body.Close()

	// Add an empty user visit.
	resp, err = http.Post(server.URL+"/users/testman/visits", "application/json", nil)
	checkErr("making http request", err)
	checkStatus("POSTing an empty visit", resp, http.StatusBadRequest)
	resp.Body.Close()

	// Get all user visits for a given user.
	resp, err = http.Get(server.URL + "/users/testman/visits")
	checkErr("making http request", err)
	checkStatus("GETing a user visit", resp, http.StatusOK)
	visitsBody := &struct {
		Visits []visits.Visit `json:"visits"`
	}{make([]visits.Visit, 0)}
	checkErr("parsing visits response body", json.NewDecoder(resp.Body).Decode(visitsBody))
	if len(visitsBody.Visits) != 1 {
		t.Fatal("expected exactly 1 visit to be returned")
	}
	if visitsBody.Visits[0].User != "testman" {
		t.Fatal("expected visit.user to be set to 'testman'")
	}
	if visitsBody.Visits[0].City != "Raleigh" {
		t.Fatal("expected visit.city to be set to 'Raleigh'")
	}
	if visitsBody.Visits[0].State != "NC" {
		t.Fatal("expected visit.state to be set to 'NC'")
	}
	raleighVisitID := visitsBody.Visits[0].ID
	resp.Body.Close()

	// Add another user visit.
	resp, err = http.Post(
		server.URL+"/users/testman/visits",
		"application/json",
		strings.NewReader(`{"city": "Charlotte", "state": "NC"}`),
	)
	checkErr("making http request", err)
	checkStatus("POSTing a valid visit", resp, http.StatusOK)
	resp.Body.Close()

	// Get all state names visited by a user.
	resp, err = http.Get(server.URL + "/users/testman/visits/states")
	checkErr("making http request", err)
	checkStatus("GETing the states a user visited", resp, http.StatusOK)
	statesBody := &struct {
		States []string `json:"states"`
	}{make([]string, 0)}
	checkErr("parsing states response body", json.NewDecoder(resp.Body).Decode(statesBody))
	if len(statesBody.States) != 1 {
		t.Fatal("expected exactly 1 unique state to be returned")
	}
	resp.Body.Close()

	// Get all city names visited by a user.
	resp, err = http.Get(server.URL + "/users/testman/visits/cities")
	checkErr("making http request", err)
	checkStatus("GETing the cities a user visited", resp, http.StatusOK)
	citiesBody := &struct {
		Cities []string `json:"cities"`
	}{make([]string, 0)}
	checkErr("parsing cities response body", json.NewDecoder(resp.Body).Decode(citiesBody))
	if len(citiesBody.Cities) != 2 {
		t.Fatal("expected exactly 2 unique cities to be returned")
	}
	resp.Body.Close()

	// Delete a user visit.
	req, err := http.NewRequest("DELETE", server.URL+"/users/testman/visits/"+raleighVisitID, nil)
	checkErr("making http request", err)
	resp, err = http.DefaultClient.Do(req)
	checkErr("failed to make http request", err)
	checkStatus("DELETEing the Raleigh user visit", resp, http.StatusNoContent)
	resp.Body.Close()

	// Get all user visits for a given user after deleting one.
	resp, err = http.Get(server.URL + "/users/testman/visits")
	checkErr("making http request", err)
	checkStatus("GETing a user visit", resp, http.StatusOK)
	visitsBody = &struct {
		Visits []visits.Visit `json:"visits"`
	}{make([]visits.Visit, 0)}
	checkErr("parsing visits response body", json.NewDecoder(resp.Body).Decode(visitsBody))
	if len(visitsBody.Visits) != 1 {
		t.Fatal("expected exactly 1 visit to be returned")
	}
	resp.Body.Close()

	// Make sure the 2 new visits were sent over the streaming endpoint.
	i := 0
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		text = text[len("data: "):]
		v := &visits.Visit{}
		checkErr("unmarshalling streamed visit json", json.Unmarshal([]byte(text), v))
		if v.State == "NC" {
			i++
		} else {
			t.Fatal("expected streamed visit.state = 'NC'")
		}
		if i == 2 {
			break
		}
	}
	resp.Body.Close()

	// Get all cities in the state of NC.
	resp, err = http.Get(server.URL + "/states/nc/cities")
	checkErr("making http request", err)
	checkStatus("GETing a list of cities in a state", resp, http.StatusOK)
	stateCitiesBody := &struct {
		Cities []string `json:"cities"`
	}{make([]string, 0)}
	checkErr("parsing cities response body", json.NewDecoder(resp.Body).Decode(stateCitiesBody))
	if len(stateCitiesBody.Cities) != 2 {
		t.Fatalf("expected exactly 2 cities to be returned, got %v", stateCitiesBody.Cities)
	}
	resp.Body.Close()
}
