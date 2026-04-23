package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"

	"github.com/go-chi/chi/v5"
)

type createTxnRequest struct {
	Amount        float64 `json:"amount"`
	SubcategoryID string  `json:"subcategoryId"`
	Type          string  `json:"type"`
	SrcID         string  `json:"srcId"`
	DstID         string  `json:"dstId"`
	ContactName   string  `json:"contactName"`
	Timestamp     int64   `json:"timestamp"`
	Remarks       string  `json:"remarks"`
}

type updateTxnRequest struct {
	Amount        float64 `json:"amount"`
	SubcategoryID string  `json:"subcategoryId"`
	Type          string  `json:"type"`
	SrcID         string  `json:"srcId"`
	DstID         string  `json:"dstId"`
	ContactName   string  `json:"contactName"`
	Timestamp     int64   `json:"timestamp"`
	Remarks       string  `json:"remarks"`
}

// HandleListTransactions handles GET /transactions.
func HandleListTransactions(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	txnType := models.TransactionType(r.URL.Query().Get("type"))
	startStr := r.URL.Query().Get("startDate")
	endStr := r.URL.Query().Get("endDate")

	start := int64(0)
	end := time.Now().Unix()
	if startStr != "" {
		start, _ = strconv.ParseInt(startStr, 10, 64)
	}
	if endStr != "" {
		end, _ = strconv.ParseInt(endStr, 10, 64)
	}

	txns, err := all.GetServices().Txn.ListTransactionsByTime(claims.UserID, txnType, start, end)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	if txns == nil {
		txns = []models.Transaction{}
	}

	// For now, return all since service doesn't support pagination yet
	// but wrap in the expected response format
	WriteJSON(w, http.StatusOK, map[string]any{
		"data": txns,
		"pagination": map[string]any{
			"page":  1,
			"limit": len(txns),
			"total": len(txns),
		},
	})
}

// HandleCreateTransaction handles POST /transactions.
func HandleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	var req createTxnRequest
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	txn := models.Transaction{
		UserID:        claims.UserID,
		Amount:        req.Amount,
		SubcategoryID: req.SubcategoryID,
		Type:          models.TransactionType(req.Type),
		SrcID:         req.SrcID,
		DstID:         req.DstID,
		ContactName:   req.ContactName,
		Timestamp:     req.Timestamp,
		Remarks:       req.Remarks,
	}

	if err := all.GetServices().Txn.AddTransaction(txn); err != nil {
		WriteError(w, http.StatusBadRequest, "create_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]string{"message": "transaction created"})
}

// HandleUpdateTransaction handles PUT /transactions/{id}.
func HandleUpdateTransaction(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	txnID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid transaction id")
		return
	}

	var req updateTxnRequest
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	txn := models.Transaction{
		Amount:        req.Amount,
		SubcategoryID: req.SubcategoryID,
		Type:          models.TransactionType(req.Type),
		SrcID:         req.SrcID,
		DstID:         req.DstID,
		ContactName:   req.ContactName,
		Timestamp:     req.Timestamp,
		Remarks:       req.Remarks,
	}

	if err := all.GetServices().Txn.UpdateTransaction(claims.UserID, txnID, txn); err != nil {
		WriteError(w, http.StatusBadRequest, "update_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "transaction updated"})
}

// HandleDeleteTransaction handles DELETE /transactions/{id}.
func HandleDeleteTransaction(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	txnID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "invalid transaction id")
		return
	}

	if err := all.GetServices().Txn.DeleteTransaction(claims.UserID, txnID); err != nil {
		WriteError(w, http.StatusBadRequest, "delete_failed", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "transaction deleted"})
}
