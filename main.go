package main

import (
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/dancannon/gorethink"
	"github.com/nstogner/beenthere-ws/cities"
	"github.com/nstogner/beenthere-ws/handler"
	"github.com/nstogner/beenthere-ws/visits"
)

var log = logrus.New()

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	// Setup DB connection/clients.
	sess, err := gorethink.Connect(gorethink.ConnectOpts{})
	if err != nil {
		log.WithField("error", err.Error()).Fatal("unable to connect to database")
	}
	vc := visits.NewClient(visits.Config{
		DB:    "been_there",
		Table: "user_visits",
	}, sess)
	cc := cities.NewClient(cities.Config{
		DB:    "been_there",
		Table: "cities",
	}, sess)

	// Setup http handler.
	hdlr := handler.New(handler.Config{
		Logger:       log,
		VisitsClient: vc,
		CitiesClient: cc,
	})
	log.WithField("port", port).Info("starting service...")
	log.Fatal(http.ListenAndServe(":"+port, hdlr))
}
