package beenthere

import (
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/dancannon/gorethink"
)

var log = logrus.New()

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	sess, err := gorethink.Connect(gorethink.ConnectOpts{})
	if err != nil {
		log.WithField("error", err.Error()).Fatal("unable to connect to database")
	}
	vc := NewVisitClient(sess)
	hdlr := NewHandler(vc)

	log.WithField("port", port).Info("starting service...")
	log.Fatal(http.ListenAndServe(":"+port, hdlr))
}
