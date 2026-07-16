package web

import (
	"net/http"

	"github.com/masudur-rahman/khorcha-pati/services/all"
)

// metaResponse is the public runtime info the SPA hydrates its config from.
type metaResponse struct {
	BotUsername string `json:"botUsername"`
}

// HandleMeta handles GET /api/v1/meta.
func HandleMeta(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, metaResponse{BotUsername: all.BotUsername()})
}
