package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	r "github.com/dancannon/gorethink"
	"github.com/nstogner/beenthere-ws/handler"
	"github.com/nstogner/beenthere-ws/locations"
	"github.com/nstogner/beenthere-ws/visits"
)

var log = logrus.New()
var session *r.Session
var config Config

func main() {
	var err error

	// Parse CLI flags.
	shouldInitDB := flag.Bool("init-db", false, "create database schema")
	flag.Parse()

	// Pull configuration from the environment.
	config = ConfigFromEnv()

	// Setup DB connection.
	session, err = r.Connect(r.ConnectOpts{
		Address:  fmt.Sprintf("%s:%s", config.DBHost, config.DBPort),
		Database: config.DBName,
	})
	if err != nil {
		log.WithField("error", err.Error()).Fatal("unable to connect to database")
	}
	defer func() {
		session.Close()
	}()

	// Decide on whether to start the server or setup the db.
	if *shouldInitDB {
		initDB()
	} else {
		runServer()
	}
}

func runServer() {
	// Setup DB clients.
	vc := visits.NewClient(visits.Config{
		Table: config.VisitsTable,
	}, session)
	lc := locations.NewClient(locations.Config{
		Table: config.CitiesTable,
	}, session)

	// Setup HTTP handler.
	hdlr := handler.New(handler.Config{
		Logger:       log,
		VisitsClient: vc,
		LocsClient:   lc,
	})
	log.WithField("port", config.ServerPort).Info("starting service...")
	log.Fatal(http.ListenAndServe(":"+config.ServerPort, hdlr))
}

func initDB() {
	log.Info("initializing database schema...")

	checkErr := func(msg string, err error) {
		if err != nil {
			log.WithError(err).Fatalf("failure: %s", msg)
		}
	}

	log.WithField("db", config.DBName).Info("creating db")
	_, err := r.DBCreate(config.DBName).RunWrite(session)
	checkErr("creating db", err)

	log.WithField("table", config.VisitsTable).Info("creating table")
	_, err = r.TableCreate(config.VisitsTable).RunWrite(session)
	checkErr("creating table", err)

	log.WithField("table", config.CitiesTable).Info("creating table")
	_, err = r.TableCreate(config.CitiesTable).RunWrite(session)
	checkErr("creating table", err)

	log.WithFields(logrus.Fields{
		"table": config.VisitsTable,
		"index": "user",
	}).Info("creating index on table")
	_, err = r.Table(config.VisitsTable).IndexCreate("user").RunWrite(session)
	checkErr("creating table index", err)

	log.WithFields(logrus.Fields{
		"table": config.CitiesTable,
		"index": "state",
	}).Info("creating index on table")
	_, err = r.Table(config.CitiesTable).IndexCreate("state").RunWrite(session)
	checkErr("creating table index", err)

	_, err = r.Table(config.VisitsTable).IndexWait("user").Run(session)
	checkErr("waiting on table index", err)
	_, err = r.Table(config.CitiesTable).IndexWait("state").Run(session)
	checkErr("waiting on table index", err)

	log.Info("successfully initialized database")
}
