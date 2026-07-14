package web

import (
	"context"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/masudur-rahman/khorcha-pati/configs"
	"github.com/masudur-rahman/khorcha-pati/infra/logr"
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

// Sort keys accepted by the admin users listing.
const (
	userSortRegistered = "registered"
	userSortLastTxn    = "last_txn"
)

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

// sortAdminUsers orders users in-place by the value returned by sortVal, descending unless desc is false.
func sortAdminUsers(users []models.Profile, sortVal func(models.Profile) int64, desc bool) {
	sort.SliceStable(users, func(i, j int) bool {
		a, b := sortVal(users[i]), sortVal(users[j])
		if desc {
			return a > b
		}
		return a < b
	})
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

// HandleAdminUsers returns all registered users with resource counts, optionally paginated.
func HandleAdminUsers(w http.ResponseWriter, r *http.Request) {
	svc := all.GetServices()
	users, err := svc.User.ListUsers()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "users_error", err.Error())
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 0
	limit := 0
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	switch r.URL.Query().Get("sort") {
	case userSortRegistered:
		desc := r.URL.Query().Get("order") != "asc"
		sortAdminUsers(users, func(u models.Profile) int64 { return u.CreatedAt }, desc)
	case userSortLastTxn:
		desc := r.URL.Query().Get("order") != "asc"
		lastByUser := make(map[int64]int64, len(users))
		for _, u := range users {
			txns, _ := svc.Txn.ListTransactions(u.ID)
			lastByUser[u.ID] = lastTxnTime(txns)
		}
		sortAdminUsers(users, func(u models.Profile) int64 { return lastByUser[u.ID] }, desc)
	}

	total := len(users)
	var paginatedUsers []models.Profile
	if page > 0 && limit > 0 {
		start := (page - 1) * limit
		end := start + limit
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}
		paginatedUsers = users[start:end]
	} else {
		paginatedUsers = users
	}

	result := make([]adminUserResponse, 0, len(paginatedUsers))
	for _, u := range paginatedUsers {
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

	WriteJSON(w, http.StatusOK, map[string]any{
		"users": result,
		"total": total,
	})
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

// HandleAdminBroadcast sends a custom message to all active users via the bot.
func HandleAdminBroadcast(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Message        string  `json:"message"`
		IncludeUserIDs []int64 `json:"includeUserIds"`
		ExcludeUserIDs []int64 `json:"excludeUserIds"`
	}
	if err := ReadJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	body.Message = strings.TrimSpace(body.Message)
	if body.Message == "" {
		WriteError(w, http.StatusBadRequest, "invalid_input", "message is required")
		return
	}

	messenger := all.GetMessenger()
	if messenger == nil {
		WriteError(w, http.StatusInternalServerError, "messenger_error", "telegram messenger is not initialized")
		return
	}

	svc := all.GetServices()
	users, err := svc.User.ListUsers()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "users_error", err.Error())
		return
	}

	includes := make(map[int64]bool)
	for _, id := range body.IncludeUserIDs {
		includes[id] = true
	}
	excludes := make(map[int64]bool)
	for _, id := range body.ExcludeUserIDs {
		excludes[id] = true
	}

	re := regexp.MustCompile(`(?i)\{\{\s*name\s*\}\}|\{\s*name\s*\}`)

	var sentCount, failCount int
	for _, u := range users {
		if u.TelegramID == 0 {
			continue
		}
		if len(includes) > 0 && !includes[u.ID] {
			continue
		}
		if excludes[u.ID] {
			continue
		}

		name := strings.TrimSpace(u.FirstName + " " + u.LastName)
		if name == "" {
			name = u.Username
		}
		if name == "" {
			name = "User"
		}

		msg := re.ReplaceAllString(body.Message, name)

		if err := messenger.SendMessage(u.TelegramID, msg); err != nil {
			logr.DefaultLogger.Errorw("failed to send broadcast message to user", "userId", u.ID, "telegramId", u.TelegramID, "error", err.Error())
			failCount++
		} else {
			sentCount++
		}
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"sent":    sentCount,
		"failed":  failCount,
	})
}
