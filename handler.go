package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
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
	visits     *VisitClient
	router     *httprouter.Router
}

// NewHandler returns an instance of Handler with registered routes.
func NewHandler(visits *VisitClient) *Handler {
	h := &Handler{
		visits: visits,
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
	rtr.GET("/states/:st/cities", h.wrap(h.GetCities))
	rtr.POST("/users/:usr/visits", h.wrap(h.PostUserVisit))
	rtr.DELETE("/users/:usr/visits/:vis", h.wrap(h.DeleteUserVisit))
	rtr.GET("/users/:usr/visits", h.wrap(h.GetUserCities))
	rtr.GET("/users/:usr/visits/states", h.wrap(h.GetUserStates))
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
	stateId := ps.ByName("st")

	_ = stateId

	return nil
}

func (h *Handler) PostUserVisit(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	userId := ps.ByName("usr")

	// Grab the visit details from the http body.
	visit := NewVisit()
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

func (h *Handler) DeleteUserVisit(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	visitId := ps.ByName("vis")

	// Delete the visit from the database.
	if err := h.visits.Delete(visitId); err != nil {
		return httpware.NewErr("unable to delete user visit", http.StatusInternalServerError).WithField("error", err.Error())
	}

	res.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) GetUserCities(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	userId := ps.ByName("usr")

	states, err := h.visits.GetFields(userId, "city")
	if err != nil {
		return httpware.NewErr(err.Error(), http.StatusInternalServerError)
	}

	rsp := contentware.ResponseTypeFromCtx(ctx)
	rsp.Encode(res, states)
	return nil
}

func (h *Handler) GetUserStates(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	userId := ps.ByName("usr")

	states, err := h.visits.GetFields(userId, "state")
	if err != nil {
		return httpware.NewErr(err.Error(), http.StatusInternalServerError)
	}

	rsp := contentware.ResponseTypeFromCtx(ctx)
	rsp.Encode(res, states)
	return nil
}
