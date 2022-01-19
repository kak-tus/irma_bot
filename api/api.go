package api

import (
	"net/http"

	oapimiddware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/getkin/kin-openapi/openapi3"
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
	log     *zap.SugaredLogger
	model   *model.Model
	router  *chi.Mux
	storage *storage.InstanceObj
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

	swagger, err := GetSwagger()
	if err != nil {
		return nil, err
	}

	// https://github.com/deepmap/oapi-codegen/blob/master/examples/petstore-expanded/echo/petstore.go#L29
	// Also add path part - this need to recognize requests
	// TODO FIX hardcode
	swagger.Servers = openapi3.Servers{
		&openapi3.Server{
			URL: "http:///api/v1",
		},
	}

	router.Route("/api/v1", func(r chi.Router) {
		r.Use(oapimiddware.OapiRequestValidator(swagger))
		_ = HandlerFromMux(hdl, r)
	})

	return hdl, nil
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
