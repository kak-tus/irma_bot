package api

import (
	"net/http"

	"github.com/goccy/go-json"
)

func (hdl *API) Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(HealthcheckResponse{})
}
