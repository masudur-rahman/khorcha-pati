package web

import (
	"net/http"

	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/services/all"
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
		WriteError(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, wallet)
}
