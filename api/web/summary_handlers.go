package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/services/all"
)

// HandleChartData handles GET /summary/charts?year=&month=&months=.
func HandleChartData(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	now := time.Now()
	year := intParam(r, "year", now.Year())
	month := intParam(r, "month", int(now.Month()))

	overview, err := all.GetServices().Summary.GetMonthlyOverview(claims.UserID, year, month)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "overview_failed", err.Error())
		return
	}

	categories, err := all.GetServices().Summary.GetExpenseByCategory(claims.UserID, year, month)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "categories_failed", err.Error())
		return
	}

	months := intParam(r, "months", 6)
	comparison, err := all.GetServices().Summary.GetIncomeVsExpense(claims.UserID, months)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "comparison_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"overview":   overview,
		"categories": categories,
		"comparison": comparison,
	})
}

// HandleListCategories handles GET /categories.
func HandleListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := all.GetServices().Txn.ListTxnCategories()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, cats)
}

// HandleGetProfile handles GET /profile.
func HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	user, err := all.GetServices().User.GetUserByID(claims.UserID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "profile_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, user)
}

// HandleUpdateProfile handles PUT /profile.
func HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	var req struct {
		MobileNumber string `json:"mobileNumber"`
		Timezone     string `json:"timezone"`
	}
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	user, err := all.GetServices().User.GetUserByID(claims.UserID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "profile_failed", err.Error())
		return
	}

	user.MobileNumber = req.MobileNumber
	user.Timezone = req.Timezone

	if err := all.GetServices().User.UpdateUser(claims.UserID, user); err != nil {
		WriteError(w, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

// HandleListSubcategories handles GET /subcategories.
func HandleListSubcategories(w http.ResponseWriter, r *http.Request) {
	catID := r.URL.Query().Get("catId")
	subs, err := all.GetServices().Txn.ListTxnSubcategories(catID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, subs)
}

func intParam(r *http.Request, key string, fallback int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return n
}
