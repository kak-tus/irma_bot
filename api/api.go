package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/goccy/go-json"
	"github.com/kak-tus/irma_bot/model"
	"github.com/kak-tus/irma_bot/storage"
	"go.uber.org/zap"
)

type Options struct {
	Log     *zap.SugaredLogger
	Model   *model.Model
	Storage *storage.InstanceObj
}

type API struct {
	httpHandler http.Handler
	log         *zap.SugaredLogger
	model       *model.Model
	router      *chi.Mux
	storage     *storage.InstanceObj
}

func NewAPI(opts Options) (*API, error) {
	router := chi.NewRouter()

	corsHdl := cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "token"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	router.Use(corsHdl)

	router.Use(middleware.SetHeader("Content-Type", "application/json"))

	hdl := &API{
		log:     opts.Log,
		model:   opts.Model,
		router:  router,
		storage: opts.Storage,
	}

	httpHdl := HandlerFromMuxWithBaseURL(hdl, router, "/api/v1")

	hdl.httpHandler = httpHdl

	return hdl, nil
}

func (hdl *API) GetHTTPHandler() http.Handler {
	return hdl.httpHandler
}

func (hdl *API) GetHTTPRouter() *chi.Mux {
	return hdl.router
}

func (hdl *API) errorNotFound(w http.ResponseWriter, err error, msg string) {
	hdl.log.Error(err)

	resp := NotFoundErrorResponse{
		Message: msg,
	}

	w.WriteHeader(http.StatusNotFound)

	_ = json.NewEncoder(w).Encode(resp)
}

func (hdl *API) errorInternal(w http.ResponseWriter, err error, msg string) {
	hdl.log.Error(err)

	resp := InternalErrorResponse{
		Message: msg,
	}

	w.WriteHeader(http.StatusInternalServerError)

	_ = json.NewEncoder(w).Encode(resp)
}
