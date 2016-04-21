package main

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/dancannon/gorethink"
	"github.com/nstogner/beenthere-ws/handler"
	"github.com/nstogner/beenthere-ws/locations"
	"github.com/nstogner/beenthere-ws/visits"
)

var log = logrus.New()

func main() {
	// Pull configuration from the environment.
	conf := ConfigFromEnv()

	// Setup DB connection/clients.
	sess, err := gorethink.Connect(gorethink.ConnectOpts{
		Address:  fmt.Sprintf("%s:%s", conf.DBHost, conf.DBPort),
		Database: conf.DBName,
	})
	if err != nil {
		log.WithField("error", err.Error()).Fatal("unable to connect to database")
	}
	vc := visits.NewClient(visits.Config{
		Table: conf.VisitsTable,
	}, sess)
	lc := locations.NewClient(locations.Config{
		Table: conf.CitiesTable,
	}, sess)

	// Setup HTTP handler.
	hdlr := handler.New(handler.Config{
		Logger:       log,
		VisitsClient: vc,
		LocsClient:   lc,
	})
	log.WithField("port", conf.ServerPort).Info("starting service...")
	log.Fatal(http.ListenAndServe(":"+conf.ServerPort, hdlr))
}
