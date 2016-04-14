package beenthere

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/nstogner/httpware"
	"github.com/nstogner/httpware/contentware"
	"github.com/nstogner/httpware/errorware"
	"github.com/nstogner/httpware/httpctx"
	"github.com/nstogner/httpware/httperr"
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
	h.middleware = httpware.MustCompose(
		contentware.New(contentware.Defaults),
		errorware.New(errorware.Defaults),
		logware.New(logware.Config{
			Logger:    log,
			Headers:   []string{},
			Successes: true,
			Failures:  true,
		}),
	)

	// Register all http routes.
	rtr := httprouter.New()
	rtr.GET("/state/:st/cities", h.wrap(h.GetCities))
	rtr.POST("/user/:usr/visits", h.wrap(h.PostUserVisit))
	rtr.DELETE("/user/:usr/visit/:vis", h.wrap(h.DeleteUserVisit))
	rtr.GET("/user/:usr/visits", h.wrap(h.GetUserCities))
	rtr.GET("/user/:usr/visits/states", h.wrap(h.GetUserStates))
	h.router = rtr

	return h
}

// wrap applies middleware to a handler function and returns a handler
// function which is compatible with httprouter.
func (h *Handler) wrap(hf httpctx.HandlerFunc) httprouter.Handle {
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
		return httperr.New("unable to parse body: "+err.Error(), http.StatusBadRequest)
	}
	if err := h.visits.Validate(visit); err != nil {
		return httperr.New("invalid visit", http.StatusBadRequest).WithField("invalid", err.Error())
	}

	// Save the visit to the database.
	if err := h.visits.AddUserVisit(userId, visit); err != nil {
		// TODO: Make sure that this wont send the error to the client.
		return httperr.New("unable to save user visit", http.StatusInternalServerError).WithField("error", err.Error())
	}

	return nil
}

func (h *Handler) DeleteUserVisit(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	userId := ps.ByName("usr")

	_ = userId

	return nil
}

func (h *Handler) GetUserCities(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	userId := ps.ByName("usr")

	_ = userId

	return nil
}

func (h *Handler) GetUserStates(ctx context.Context, res http.ResponseWriter, req *http.Request) error {
	ps := routeradapt.ParamsFromCtx(ctx)
	userId := ps.ByName("usr")

	_ = userId

	return nil
}
