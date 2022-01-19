package api

import (
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

func (hdl *API) GetTokenData(w http.ResponseWriter, r *http.Request, params GetTokenDataParams) {
	data, err := hdl.storage.GetTokenData(r.Context(), string(params.Token))
	if err != nil {
		hdl.errorInternal(w, err, "internal error")
		return
	}

	if data.ChatID == 0 {
		hdl.errorNotFound(w, err, "not found")
		return
	}

	resp := GetTokenResponse{
		Ttl: time.Now().Add(data.TTL).Format(time.RFC3339),
	}

	_ = json.NewEncoder(w).Encode(resp)
}
