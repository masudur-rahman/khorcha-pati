package web

import (
	"net/http"
	"strconv"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/services/all"

	"github.com/go-chi/chi/v5"
)

// HandleListContacts handles GET /contacts.
func HandleListContacts(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	contacts, err := all.GetServices().Contact.ListContacts(claims.UserID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	// Back-compat: only wrap in the paginated envelope when a limit is requested;
	// otherwise return the legacy bare array the frontend still reads.
	page, limit := parsePageLimit(r)
	if limit > 0 {
		items, total := slicePage(contacts, page, limit)
		writePaged(w, http.StatusOK, items, page, limit, total)
		return
	}
	WriteJSON(w, http.StatusOK, contacts)
}

// HandleCreateContact handles POST /contacts.
func HandleCreateContact(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	var req struct {
		NickName    string `json:"nickName"`
		FullName    string `json:"fullName"`
		Email       string `json:"email"`
		ContactInfo string `json:"contactInfo"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	contact := &models.Contacts{
		UserID:      claims.UserID,
		NickName:    req.NickName,
		FullName:    req.FullName,
		Email:       req.Email,
		ContactInfo: req.ContactInfo,
	}

	if err := all.GetServices().Contact.CreateContact(contact); err != nil {
		status, msg := models.ParseStatusError(err)
		WriteError(w, status, "create_failed", msg)
		return
	}

	WriteJSON(w, http.StatusCreated, contact)
}

// HandleUpdateContact handles PUT /contacts/{id}.
func HandleUpdateContact(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid contact id")
		return
	}

	var req struct {
		NickName    string `json:"nickName"`
		FullName    string `json:"fullName"`
		Email       string `json:"email"`
		ContactInfo string `json:"contactInfo"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	if err := all.GetServices().Contact.UpdateContact(claims.UserID, id, req.NickName, req.FullName, req.Email); err != nil {
		status, msg := models.ParseStatusError(err)
		WriteError(w, status, "update_failed", msg)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "contact updated"})
}

// HandleDeleteContact handles DELETE /contacts/{id}.
func HandleDeleteContact(w http.ResponseWriter, r *http.Request) {
	_, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid contact id")
		return
	}

	if err := all.GetServices().Contact.DeleteContact(id); err != nil {
		status, msg := models.ParseStatusError(err)
		WriteError(w, status, "delete_failed", msg)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "contact deleted"})
}
