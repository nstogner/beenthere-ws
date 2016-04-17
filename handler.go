package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/beenthere-ws/cities"
	"github.com/nstogner/beenthere-ws/visits"
	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/contentware"
	"github.com/nstogner/httpware/logware"
	"github.com/nstogner/httpware/routeradapt"
	"golang.org/x/net/context"
)

// Handler maintains a database clients and fulfills the http.Handler
// interface.
type Handler struct {
	middleware *httpware.Composite
	visits     *visits.Client
	cities     *cities.Client
	router     *httprouter.Router
}

// NewHandler returns an instance of Handler with registered routes.
func NewHandler(visits *visits.Client, locs *cities.Client) *Handler {
	h := &Handler{
		visits: visits,
		cities: locs,
	}

	// Configure any needed middleware.
	h.middleware = httpware.Compose(
		contentware.New(contentware.Defaults),
		logware.New(logware.Config{
			Logger:    log,
			Headers:   []string{},
			Successes: true,
			Failures:  true,
		}),
	)

	// Register all http routes. Note: plural names are used to adhere with
	// RESTful conventions.
	rtr := httprouter.New()
	rtr.GET("/states/:state/cities", h.wrap(h.GetCities))
	rtr.POST("/users/:user/visits", h.wrap(h.PostUserVisit))
	rtr.DELETE("/users/:user/visits/:visit", h.wrap(h.DeleteVisit))
	rtr.GET("/users/:user/visits", h.wrap(h.GetVisits))
	rtr.GET("/users/:user/visits/cities", h.wrap(h.GetCitiesVisited))
	rtr.GET("/users/:user/visits/states", h.wrap(h.GetStatesVisited))
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

func (h *Handler) GetCities(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	state := ps.ByName("state")

	stateName := h.cities.StateName(state)
	if stateName == "" {
		return httpware.NewErr("no such state", http.StatusBadRequest)
	}

	// TODO: Return all cities in a state.

	return nil
}

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

	// Save the visit to the database.
	if err := h.visits.Add(visit); err != nil {
		return httpware.NewErr("unable to save user visit", http.StatusInternalServerError).WithField("error", err.Error())
	}

	// Pass the saved entity back to the client.
	rst := contentware.ResponseTypeFromCtx(ctx)
	rst.Encode(res, visit)

	return nil
}

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

func (h *Handler) GetVisits(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	userId := ps.ByName("user")

	cities, err := h.visits.GetVisits(userId)
	if err != nil {
		return httpware.NewErr(err.Error(), http.StatusInternalServerError)
	}

	rsp := contentware.ResponseTypeFromCtx(ctx)
	rsp.Encode(res, cities)
	return nil
}

func (h *Handler) GetCitiesVisited(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	userId := ps.ByName("user")

	cities, err := h.visits.GetCities(userId)
	if err != nil {
		return httpware.NewErr(err.Error(), http.StatusInternalServerError)
	}

	rsp := contentware.ResponseTypeFromCtx(ctx)
	rsp.Encode(res, cities)
	return nil
}

func (h *Handler) GetStatesVisited(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	userId := ps.ByName("user")

	states, err := h.visits.GetStates(userId)
	if err != nil {
		return httpware.NewErr(err.Error(), http.StatusInternalServerError)
	}

	rsp := contentware.ResponseTypeFromCtx(ctx)
	rsp.Encode(res, states)
	return nil
}
