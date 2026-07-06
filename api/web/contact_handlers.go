package web

import (
	"net/http"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/services/all"
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
		WriteError(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, contact)
}
