package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/masudur-rahman/khorcha-pati/api/handlers"
	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/models/gqtypes"
	"github.com/masudur-rahman/khorcha-pati/modules/convert"
	"github.com/masudur-rahman/khorcha-pati/pkg"
	"github.com/masudur-rahman/khorcha-pati/services/all"
)

// clampStartTime returns the effective start date for a report period.
// Fallback chain: registration time → earliest txn time → startTime (computed).
func clampStartTime(startTime time.Time, registeredAt int64, txns []models.Transaction) time.Time {
	if registeredAt != 0 {
		if reg := time.Unix(registeredAt, 0); reg.After(startTime) {
			return reg
		}
		return startTime
	}
	// fallback: earliest transaction timestamp
	if len(txns) > 0 {
		earliest := txns[0].Timestamp
		for _, t := range txns[1:] {
			if t.Timestamp < earliest {
				earliest = t.Timestamp
			}
		}
		if first := time.Unix(earliest, 0); first.After(startTime) {
			return first
		}
	}
	return startTime
}

// parseTimeRange extracts the start and end time from the request.
// It supports explicit "start" and "end" query parameters, or falls back to "duration".
func parseTimeRange(r *http.Request, loc *time.Location) (time.Time, time.Time, error) {
	if loc == nil {
		loc = pkg.DefaultLocation
	}
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr != "" && endStr != "" {
		startTime, err := time.ParseInLocation("2006-01-02", startStr, loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		endTime, err := time.ParseInLocation("2006-01-02", endStr, loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		// Include the entire end day
		endTime = endTime.Add(24*time.Hour - time.Second)
		return startTime, endTime, nil
	}

	duration := handlers.SummaryDuration(r.URL.Query().Get("duration"))
	if duration == "" {
		duration = handlers.DurationThisMonth
	}
	now := time.Now().In(loc)
	startTime := handlers.CalculateStartTime(duration, loc)
	return startTime, now, nil
}

// HandleGetReport handles GET /summary/report.
func HandleGetReport(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	svc := all.GetServices()
	user, err := svc.User.GetUserByID(claims.UserID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "profile_failed", err.Error())
		return
	}

	tz := pkg.LoadTimezone(user.Timezone)
	startTime, now, err := parseTimeRange(r, tz)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_date", err.Error())
		return
	}

	txns, err := svc.Txn.ListTransactionsByTime(user.ID, "", startTime.Unix(), now.Unix())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	startTime = clampStartTime(startTime, user.CreatedAt, txns)

	report := gqtypes.Report{
		Name:      fmt.Sprintf("%v %v", user.FirstName, user.LastName),
		StartDate: startTime,
		EndDate:   now,
	}

	wallets, err := svc.Wallet.ListWallets(user.ID)
	if err == nil {
		report.Wallets = make([]gqtypes.Wallet, 0, len(wallets))
		for _, w := range wallets {
			report.Wallets = append(report.Wallets, convert.ToWalletAPIFormat(w))
		}
	}

	contacts, err := svc.Contact.ListContacts(user.ID)
	if err == nil {
		report.Contacts = make([]gqtypes.Contact, 0, len(contacts))
		for _, c := range contacts {
			report.Contacts = append(report.Contacts, convert.ToContactAPIFormat(c))
		}
	}

	txnApis := make([]gqtypes.Transaction, 0, len(txns))
	for _, txn := range txns {
		txnApis = append(txnApis, convert.ToTransactionAPIFormat(txn))
	}
	report.Transactions = txnApis

	summary, err := handlers.BuildSummary(svc, txns)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "summary_failed", err.Error())
		return
	}
	report.Summary = summary

	report.TypeSummary = gqtypes.SortMapToSlice(summary.Type)
	report.CategorySummary, err = handlers.BuildTypeSeparatedSummary(svc, txns, handlers.CategoryKeyFn, svc.Txn.GetTxnCategoryName)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "cat_summary_failed", err.Error())
		return
	}
	report.SubcategorySummary, err = handlers.BuildTypeSeparatedSummary(svc, txns, handlers.SubcategoryKeyFn, svc.Txn.GetTxnSubcategoryName)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "subcat_summary_failed", err.Error())
		return
	}

	report.TotalAmount, report.NetBalance = handlers.ComputeTotals(txns)
	handlers.FinalizeReportTxns(&report, now)

	pdfFile, err := handlers.GenerateTransactionStatementFromTemplate(report, "Khorcha-Pati Statement")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "pdf_failed", err.Error())
		return
	}
	defer os.Remove(pdfFile)

	startFmt := startTime.Format("Jan-2006")
	endFmt := now.Format("Jan-2006")
	filename := fmt.Sprintf("khorcha-pati-statement-%s-to-%s.pdf", startFmt, endFmt)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	http.ServeFile(w, r, pdfFile)
}

// HandleGetReportData handles GET /summary/report-data.
// Returns the raw report data as JSON for client-side rendering.
func HandleGetReportData(w http.ResponseWriter, r *http.Request) {
	claims, ok := UserFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}

	svc := all.GetServices()
	user, err := svc.User.GetUserByID(claims.UserID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "profile_failed", err.Error())
		return
	}

	tz := pkg.LoadTimezone(user.Timezone)
	startTime, now, err := parseTimeRange(r, tz)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_date", err.Error())
		return
	}

	txns, err := svc.Txn.ListTransactionsByTime(user.ID, "", startTime.Unix(), now.Unix())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	startTime = clampStartTime(startTime, user.CreatedAt, txns)

	report := gqtypes.Report{
		Name:      fmt.Sprintf("%v %v", user.FirstName, user.LastName),
		StartDate: startTime,
		EndDate:   now,
	}

	wallets, err := svc.Wallet.ListWallets(user.ID)
	if err == nil {
		report.Wallets = make([]gqtypes.Wallet, 0, len(wallets))
		for _, wl := range wallets {
			report.Wallets = append(report.Wallets, convert.ToWalletAPIFormat(wl))
		}
	}

	contacts, err := svc.Contact.ListContacts(user.ID)
	if err == nil {
		report.Contacts = make([]gqtypes.Contact, 0, len(contacts))
		for _, c := range contacts {
			report.Contacts = append(report.Contacts, convert.ToContactAPIFormat(c))
		}
	}

	txnApis := make([]gqtypes.Transaction, 0, len(txns))
	for _, txn := range txns {
		txnApis = append(txnApis, convert.ToTransactionAPIFormat(txn))
	}
	report.Transactions = txnApis

	summary, err := handlers.BuildSummary(svc, txns)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "summary_failed", err.Error())
		return
	}
	report.Summary = summary

	report.TypeSummary = gqtypes.SortMapToSlice(summary.Type)
	report.CategorySummary, err = handlers.BuildTypeSeparatedSummary(svc, txns, handlers.CategoryKeyFn, svc.Txn.GetTxnCategoryName)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "cat_summary_failed", err.Error())
		return
	}
	report.SubcategorySummary, err = handlers.BuildTypeSeparatedSummary(svc, txns, handlers.SubcategoryKeyFn, svc.Txn.GetTxnSubcategoryName)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "subcat_summary_failed", err.Error())
		return
	}

	report.TotalAmount, report.NetBalance = handlers.ComputeTotals(txns)
	handlers.FinalizeReportTxns(&report, now)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}
