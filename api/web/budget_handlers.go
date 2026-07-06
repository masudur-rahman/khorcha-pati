package web

import (
	"net/http"

	"github.com/masudur-rahman/khorcha-pati/services/all"

	"github.com/go-chi/chi/v5"
)

type setBudgetRequest struct {
	CategoryID string  `json:"categoryId"`
	Amount     float64 `json:"amount"`
	AlertAt    int64   `json:"alertAt"`
}

// HandleListBudgets handles GET /budgets.
func HandleListBudgets(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	statuses, err := all.GetServices().Budget.ListBudgetStatuses(claims.UserID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, statuses)
}

// HandleSetBudget handles POST /budgets.
func HandleSetBudget(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	var req setBudgetRequest
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	if req.Amount <= 0 {
		WriteError(w, http.StatusBadRequest, "bad_request", "positive amount required")
		return
	}

	if err := all.GetServices().Budget.SetBudget(claims.UserID, req.CategoryID, req.Amount, req.AlertAt); err != nil {
		WriteError(w, http.StatusBadRequest, "set_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "budget set"})
}

// HandleDeleteBudget handles DELETE /budgets/{categoryID}.
func HandleDeleteBudget(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	catID := chi.URLParam(r, "categoryID")
	if catID == "" {
		WriteError(w, http.StatusBadRequest, "bad_request", "categoryID required")
		return
	}

	if err := all.GetServices().Budget.DeleteBudget(claims.UserID, catID); err != nil {
		WriteError(w, http.StatusBadRequest, "delete_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "budget deleted"})
}

// HandleBudgetAlerts handles GET /budgets/alerts.
func HandleBudgetAlerts(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	alerts, err := all.GetServices().Budget.ListAllBudgetAlerts(claims.UserID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "alerts_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, alerts)
}
