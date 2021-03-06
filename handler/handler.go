package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/beenthere-ws/locations"
	"github.com/nstogner/beenthere-ws/visits"
	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/contentware"
	"github.com/nstogner/httpware/logware"
	"github.com/nstogner/httpware/pageware"
	"github.com/nstogner/httpware/routeradapt"
	"github.com/nstogner/httpware/streamware"
	"golang.org/x/net/context"
)

// Handler maintains a database clients and fulfills the http.Handler
// interface.
type Handler struct {
	middleware *httpware.Composite
	visits     *visits.Client
	locations  *locations.Client
	router     *httprouter.Router
	logger     *logrus.Logger
}

// Config is used to create a new instance of Handler in New(...).
type Config struct {
	Logger       *logrus.Logger
	VisitsClient *visits.Client
	LocsClient   *locations.Client
}

// New returns an instance of Handler with registered routes.
func New(conf Config) *Handler {
	h := &Handler{
		logger:    conf.Logger,
		visits:    conf.VisitsClient,
		locations: conf.LocsClient,
	}

	// Configure any needed middleware.
	h.middleware = httpware.Compose(
		httpware.DefaultErrHandler,
		contentware.New(contentware.Defaults),
		logware.New(logware.Config{
			Logger: h.logger,
		}),
	)
	paginated := h.middleware.With(
		pageware.New(pageware.Defaults),
	)
	streaming := h.middleware.With(
		streamware.New(streamware.Defaults),
	)

	// Register all http routes. Note: plural names are used to adhere with
	// RESTful conventions.
	rtr := httprouter.New()
	rtr.GET("/states/:state/cities", h.wrap(h.GetCities))
	rtr.POST("/users/:user/visits", h.wrap(h.PostUserVisit))
	rtr.DELETE("/users/:user/visits/:visit", h.wrap(h.DeleteVisit))
	// Paginate the visits endpoint.
	rtr.GET(
		"/users/:user/visits",
		routeradapt.Adapt(paginated.ThenFunc(h.GetVisits)),
	)
	rtr.GET("/users/:user/visits/cities", h.wrap(h.GetCitiesVisited))
	rtr.GET("/users/:user/visits/states", h.wrap(h.GetStatesVisited))
	rtr.GET(
		"/stream/visits",
		routeradapt.Adapt(streaming.ThenFunc(h.StreamVisits)),
	)
	h.router = rtr

	return h
}

// wrap applies middleware to a handler function and returns a handler
// function which is compatible with httprouter.
func (h *Handler) wrap(hf httpware.HandlerFunc) httprouter.Handle {
	return routeradapt.Adapt(h.middleware.ThenFunc(hf))
}

// ServeHTTP fulfills the http.Handler interface.
func (h *Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	h.router.ServeHTTP(res, req)
}

// GetCities serves a list of cities in a given state.
func (h *Handler) GetCities(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	state := ps.ByName("state")

	stateName := h.locations.StateName(state)
	if stateName == "" {
		return httpware.NewErr("no such state", http.StatusNotFound)
	}

	dbCities, err := h.locations.GetCityNames(state)
	if err != nil {
		return httpware.NewErr(err.Error(), http.StatusInternalServerError)
	}

	rst := contentware.ResponseTypeFromCtx(ctx)
	rst.Encode(res, struct {
		Cities []string `json:"cities" xml:"cities"`
	}{dbCities})
	return nil
}

// PostUserVisit adds a city/state that a user has visited.
func (h *Handler) PostUserVisit(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	userId := ps.ByName("user")

	// Grab the visit details from the http body.
	visit := visits.NewVisit()
	rqt := contentware.RequestTypeFromCtx(ctx)
	if err := rqt.Decode(req.Body, visit); err != nil {
		return httpware.NewErr("unable to parse body: "+err.Error(), http.StatusBadRequest)
	}
	if err := h.visits.Validate(visit); err != nil {
		return httpware.NewErr("invalid visit", http.StatusBadRequest).WithField("invalid", err.Error())
	}
	visit.User = userId

	// Check and see if the given State exists.
	// TODO: How should City verification work?
	//       Should a new visit be rejected if the given city doesnt exist in db?
	//       Should unknown cities be accepted and verified offline?
	city := locations.CityFromVisit(visit)
	if err := h.locations.ValidateCity(city); err != nil {
		return httpware.NewErr("invalid visit", http.StatusBadRequest).WithField("invalid", err.Error())
	}

	// Save the visit to the database.
	if err := h.visits.Add(visit); err != nil {
		return httpware.NewErr("unable to save user visit", http.StatusInternalServerError).WithField("error", err.Error())
	}

	// Pass the saved entity back to the client.
	rst := contentware.ResponseTypeFromCtx(ctx)
	rst.Encode(res, visit)

	return nil
}

// DeleteVisit removes a given user's previously added visit.
func (h *Handler) DeleteVisit(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	visitId := ps.ByName("visit")

	// Delete the visit from the database.
	if err := h.visits.Delete(visitId); err != nil {
		return httpware.NewErr("unable to delete user visit", http.StatusInternalServerError).WithField("error", err.Error())
	}

	res.WriteHeader(http.StatusNoContent)
	return nil
}

// GetVisits serves a list of visit info for a given user.
func (h *Handler) GetVisits(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	userId := ps.ByName("user")
	page := pageware.PageFromCtx(ctx)

	dbVisits, err := h.visits.GetVisits(userId, page.Start, page.Limit)
	if err != nil {
		return httpware.NewErr(err.Error(), http.StatusInternalServerError)
	}

	rsp := contentware.ResponseTypeFromCtx(ctx)
	rsp.Encode(res, struct {
		Visits []visits.Visit `json:"visits" xml:"visits"`
	}{dbVisits})
	return nil
}

// GetCitiesVisited serves a unique list of cities that have been visited by a
// given user.
func (h *Handler) GetCitiesVisited(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	userId := ps.ByName("user")

	// Grab a unique list of cities visited by the given user.
	dbCities, err := h.visits.GetCities(userId)
	if err != nil {
		return httpware.NewErr(err.Error(), http.StatusInternalServerError)
	}

	rsp := contentware.ResponseTypeFromCtx(ctx)
	rsp.Encode(res, struct {
		Cities []string `json:"cities" xml:"cities"`
	}{dbCities})
	return nil
}

// GetStatesVisited serves a unique list of states that have been visited by a
// given user.
func (h *Handler) GetStatesVisited(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	userId := ps.ByName("user")

	// Grab a unique list of states visited by the given user.
	dbStates, err := h.visits.GetStates(userId)
	if err != nil {
		return httpware.NewErr(err.Error(), http.StatusInternalServerError)
	}
	// Map state abbreviations to names.
	for i, s := range dbStates {
		dbStates[i] = h.locations.StateName(s)
	}

	rsp := contentware.ResponseTypeFromCtx(ctx)
	rsp.Encode(res, struct {
		States []string `json:"states" xml:"states"`
	}{dbStates})
	return nil
}

// StreamVisits opens a connection for sending live user visit updates via
// Server Sent Events (SSE).
func (h *Handler) StreamVisits(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	sender := streamware.SenderFromCtx(ctx)
	stream, err := h.visits.Stream()
	if err != nil {
		return httpware.NewErr(err.Error(), http.StatusInternalServerError)
	}
	visit := &visits.Visit{}
	for stream.Next(visit) {
		js, err := json.Marshal(visit)
		if err != nil {
			return httpware.NewErr("unable to marshal visit into json: "+err.Error(), http.StatusInternalServerError)
		}
		sender.Send(string(js))
		visit = &visits.Visit{}
	}
	return nil
}
