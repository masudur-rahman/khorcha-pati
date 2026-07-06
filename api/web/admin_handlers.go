package web

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/masudur-rahman/khorcha-pati/configs"
	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/services/all"

	"github.com/go-chi/chi/v5"
)

type adminStatsResponse struct {
	UserCount    int    `json:"userCount"`
	TxnCount     int    `json:"txnCount"`
	WalletCount  int    `json:"walletCount"`
	DatabaseType string `json:"databaseType"`
}

type adminUserResponse struct {
	ID           int64  `json:"id"`
	TelegramID   int64  `json:"telegramId"`
	Username     string `json:"username"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	IsAdmin      bool   `json:"isAdmin"`
	IsActive     bool   `json:"isActive"`
	WalletCount  int    `json:"walletCount"`
	TxnCount     int    `json:"txnCount"`
	ContactCount int    `json:"contactCount"`
	CreatedAt    int64  `json:"createdAt"`
	LastTxnAt    int64  `json:"lastTxnAt"`
}

// lastTxnTime returns the most recent CreatedAt across a user's transactions, or 0 if none.
func lastTxnTime(txns []models.Transaction) int64 {
	var last int64
	for _, t := range txns {
		if t.CreatedAt > last {
			last = t.CreatedAt
		}
	}
	return last
}

// HandleAdminStats returns system-wide statistics.
func HandleAdminStats(w http.ResponseWriter, _ *http.Request) {
	svc := all.GetServices()
	users, err := svc.User.ListUsers()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "stats_error", err.Error())
		return
	}

	txnCount, walletCount := countResources()

	dbType := string(configs.TrackerConfig.Database.Type)
	if dbType == "" {
		dbType = "sqlite"
	}

	WriteJSON(w, http.StatusOK, adminStatsResponse{
		UserCount:    len(users),
		TxnCount:     int(txnCount),
		WalletCount:  int(walletCount),
		DatabaseType: dbType,
	})
}

// HandleAdminUsers returns all registered users with resource counts.
func HandleAdminUsers(w http.ResponseWriter, _ *http.Request) {
	svc := all.GetServices()
	users, err := svc.User.ListUsers()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "users_error", err.Error())
		return
	}

	result := make([]adminUserResponse, 0, len(users))
	for _, u := range users {
		wallets, _ := svc.Wallet.ListWallets(u.ID)
		txns, _ := svc.Txn.ListTransactions(u.ID)
		contacts, _ := svc.Contact.ListContacts(u.ID)
		result = append(result, adminUserResponse{
			ID:           u.ID,
			TelegramID:   u.TelegramID,
			Username:     u.Username,
			FirstName:    u.FirstName,
			LastName:     u.LastName,
			IsAdmin:      u.IsAdmin,
			IsActive:     u.IsActive,
			WalletCount:  len(wallets),
			TxnCount:     len(txns),
			ContactCount: len(contacts),
			CreatedAt:    u.CreatedAt,
			LastTxnAt:    lastTxnTime(txns),
		})
	}

	WriteJSON(w, http.StatusOK, result)
}

// HandleAdminUserDetail returns detail for a single user by ID.
func HandleAdminUserDetail(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "user ID must be a number")
		return
	}

	svc := all.GetServices()
	user, err := svc.User.GetUserByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "user_error", err.Error())
		return
	}

	wallets, _ := svc.Wallet.ListWallets(user.ID)
	txns, _ := svc.Txn.ListTransactions(user.ID)
	contacts, _ := svc.Contact.ListContacts(user.ID)

	WriteJSON(w, http.StatusOK, adminUserResponse{
		ID:           user.ID,
		TelegramID:   user.TelegramID,
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		IsAdmin:      user.IsAdmin,
		IsActive:     user.IsActive,
		WalletCount:  len(wallets),
		TxnCount:     len(txns),
		ContactCount: len(contacts),
		CreatedAt:    user.CreatedAt,
		LastTxnAt:    lastTxnTime(txns),
	})
}

// HandleAdminSetUserActive enables or disables a user. Body: {"isActive": bool}.
// Returns 400 if the admin tries to disable themselves.
func HandleAdminSetUserActive(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "user ID must be a number")
		return
	}

	if id == claims.UserID {
		WriteError(w, http.StatusBadRequest, "self_action", "cannot change your own active status")
		return
	}

	var body struct {
		IsActive bool `json:"isActive"`
	}
	if err := ReadJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	svc := all.GetServices()
	if _, err := svc.User.GetUserByID(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "user_error", err.Error())
		return
	}

	if err := svc.User.SetActive(id, body.IsActive); err != nil {
		WriteError(w, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{"id": id, "isActive": body.IsActive})
}

func countResources() (txnCount, walletCount int64) {
	db := configs.GetUnitOfWork().SQL
	bgCtx := context.Background()

	var txns []models.Transaction
	if err := db.Table(models.Transaction{}.TableName()).FindMany(bgCtx, &txns); err == nil {
		txnCount = int64(len(txns))
	}

	var ws []models.Wallet
	if err := db.Table(models.Wallet{}.TableName()).FindMany(bgCtx, &ws); err == nil {
		walletCount = int64(len(ws))
	}

	return txnCount, walletCount
}
