package web

import (
	"net/http"
	"strconv"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/services/all"

	"github.com/go-chi/chi/v5"
)

type accessSettingsResponse struct {
	Restricted   bool                 `json:"restricted"`
	ReplyText    string               `json:"replyText"`
	AllowedUsers []models.AllowedUser `json:"allowedUsers"`
}

func writeAccessSettings(w http.ResponseWriter, includeRevoked bool) {
	acc := all.GetServices().Access
	WriteJSON(w, http.StatusOK, accessSettingsResponse{
		Restricted:   acc.IsRestricted(),
		ReplyText:    acc.RestrictedReplyText(),
		AllowedUsers: acc.ListAllowedUsers(includeRevoked),
	})
}

// HandleGetAccessSettings handles GET /admin/access?all=true.
func HandleGetAccessSettings(w http.ResponseWriter, r *http.Request) {
	writeAccessSettings(w, r.URL.Query().Get("all") == "true")
}

// HandleUpdateAccessSettings handles PUT /admin/access.
func HandleUpdateAccessSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Restricted *bool   `json:"restricted"`
		ReplyText  *string `json:"replyText"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	acc := all.GetServices().Access
	if req.Restricted != nil {
		if err := acc.SetRestricted(*req.Restricted); err != nil {
			WriteError(w, http.StatusInternalServerError, "update_failed", err.Error())
			return
		}
	}
	if req.ReplyText != nil {
		if err := acc.SetReplyText(*req.ReplyText); err != nil {
			WriteError(w, http.StatusInternalServerError, "update_failed", err.Error())
			return
		}
	}
	writeAccessSettings(w, false)
}

// HandleAddAllowedUser handles POST /admin/access/allowed.
func HandleAddAllowedUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username   string `json:"username"`
		TelegramID int64  `json:"telegramId"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	entry, err := all.GetServices().Access.Allow(req.Username, req.TelegramID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "allow_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, entry)
}

// HandleRemoveAllowedUser handles DELETE /admin/access/allowed/{id} — a soft
// revoke: the row is tombstoned, not deleted.
func HandleRemoveAllowedUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	if err := all.GetServices().Access.Revoke(id); err != nil {
		WriteError(w, http.StatusInternalServerError, "revoke_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"id": id, "revoked": true})
}

// HandleRestoreAllowedUser handles POST /admin/access/allowed/{id}/restore.
func HandleRestoreAllowedUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	if err := all.GetServices().Access.Restore(id); err != nil {
		WriteError(w, http.StatusInternalServerError, "restore_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"id": id, "revoked": false})
}
