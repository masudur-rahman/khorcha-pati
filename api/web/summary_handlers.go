package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"
)

type categoryWithTypes struct {
	models.TxnCategory
	Types []models.TransactionType `json:"types"`
}

type subcategoryWithTypes struct {
	models.TxnSubcategory
	Types []models.TransactionType `json:"types"`
}

// parseTxnTypeQuery reads the optional ?type= filter and validates it.
func parseTxnTypeQuery(r *http.Request) (models.TransactionType, bool, error) {
	raw := r.URL.Query().Get("type")
	if raw == "" {
		return "", false, nil
	}
	typ := models.TransactionType(raw)
	switch typ {
	case models.ExpenseTransaction, models.IncomeTransaction, models.TransferTransaction:
		return typ, true, nil
	}
	return "", false, errUnknownTxnType
}

var errUnknownTxnType = &httpError{Status: http.StatusBadRequest, Code: "invalid_type", Message: "type must be Expense, Income or Transfer"}

type httpError struct {
	Status  int
	Code    string
	Message string
}

func (e *httpError) Error() string { return e.Message }

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

// HandleListCategories handles GET /categories?type=Expense|Income|Transfer (type optional).
func HandleListCategories(w http.ResponseWriter, r *http.Request) {
	typ, hasFilter, err := parseTxnTypeQuery(r)
	if err != nil {
		he := err.(*httpError)
		WriteError(w, he.Status, he.Code, he.Message)
		return
	}

	cats, err := all.GetServices().Txn.ListTxnCategories()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}

	out := make([]categoryWithTypes, 0, len(cats))
	for _, cat := range cats {
		types := models.CategoryTypes[cat.ID]
		if hasFilter && !models.ContainsType(types, typ) {
			continue
		}
		out = append(out, categoryWithTypes{TxnCategory: cat, Types: types})
	}
	WriteJSON(w, http.StatusOK, out)
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

// HandleListSubcategories handles GET /subcategories?catId=&type=Expense|Income|Transfer (both optional).
func HandleListSubcategories(w http.ResponseWriter, r *http.Request) {
	typ, hasFilter, err := parseTxnTypeQuery(r)
	if err != nil {
		he := err.(*httpError)
		WriteError(w, he.Status, he.Code, he.Message)
		return
	}

	catID := r.URL.Query().Get("catId")
	subs, err := all.GetServices().Txn.ListTxnSubcategories(catID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}

	out := make([]subcategoryWithTypes, 0, len(subs))
	for _, sub := range subs {
		types := models.SubcategoryTypes[sub.ID]
		if hasFilter && !models.ContainsType(types, typ) {
			continue
		}
		out = append(out, subcategoryWithTypes{TxnSubcategory: sub, Types: types})
	}
	WriteJSON(w, http.StatusOK, out)
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
