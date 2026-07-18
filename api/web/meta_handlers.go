package web

import (
	"net/http"
	"strings"

	"github.com/masudur-rahman/khorcha-pati/configs"
	"github.com/masudur-rahman/khorcha-pati/services/all"
)

// metaResponse is the public runtime info the SPA hydrates its config from.
type metaResponse struct {
	BotUsername string `json:"botUsername"`
	// BotOwner is the admin's Telegram username — the "contact for help" target.
	BotOwner string `json:"botOwner"`
}

// HandleMeta handles GET /api/v1/meta.
func HandleMeta(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, metaResponse{
		BotUsername: all.BotUsername(),
		BotOwner:    strings.TrimPrefix(configs.TrackerConfig.Telegram.BotOwner, "@"),
	})
}
