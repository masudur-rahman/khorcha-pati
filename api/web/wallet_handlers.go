package web

import (
	"net/http"
	"strconv"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/services/all"

	"github.com/go-chi/chi/v5"
)

// HandleListWallets handles GET /wallets.
func HandleListWallets(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	wallets, err := all.GetServices().Wallet.ListWallets(claims.UserID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, wallets)
}

// HandleCreateWallet handles POST /wallets.
func HandleCreateWallet(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	var req struct {
		Type      string  `json:"type"`
		ShortName string  `json:"shortName"`
		Name      string  `json:"name"`
		Balance   float64 `json:"balance"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	wallet := &models.Wallet{
		UserID:    claims.UserID,
		Type:      models.WalletType(req.Type),
		ShortName: req.ShortName,
		Name:      req.Name,
		Balance:   req.Balance,
	}

	if err := all.GetServices().Wallet.CreateWallet(wallet); err != nil {
		status, msg := models.ParseStatusError(err)
		WriteError(w, status, "create_failed", msg)
		return
	}

	WriteJSON(w, http.StatusCreated, wallet)
}

// HandleUpdateWallet handles PUT /wallets/{id}.
func HandleUpdateWallet(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid wallet id")
		return
	}

	var req struct {
		Name      string `json:"name"`
		ShortName string `json:"shortName"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	if err := all.GetServices().Wallet.UpdateWallet(claims.UserID, id, req.Name, req.ShortName); err != nil {
		status, msg := models.ParseStatusError(err)
		WriteError(w, status, "update_failed", msg)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "wallet updated"})
}

// HandleDeleteWallet handles DELETE /wallets/{shortName}.
func HandleDeleteWallet(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	shortName := chi.URLParam(r, "shortName")
	if shortName == "" {
		WriteError(w, http.StatusBadRequest, "bad_request", "missing wallet shortName")
		return
	}

	if err := all.GetServices().Wallet.DeleteWallet(claims.UserID, shortName); err != nil {
		status, msg := models.ParseStatusError(err)
		WriteError(w, status, "delete_failed", msg)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "wallet deleted"})
}
