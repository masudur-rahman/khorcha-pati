package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/configs"
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/models/gqtypes"
	"github.com/masudur-rahman/expense-tracker-bot/modules/convert"
	"github.com/masudur-rahman/expense-tracker-bot/pkg"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"
	"github.com/masudur-rahman/expense-tracker-bot/templates"

	"gopkg.in/telebot.v3"
)

type ReportCallbackOptions struct {
	Duration SummaryDuration `json:"duration"`
}

func TransactionReportCallback(ctx telebot.Context) error {
	callbackOpts := CallbackOptions{
		Type: ReportTypeCallback,
	}

	inlineButtons := generateReportDurationInlineButton(callbackOpts)

	return ctx.Send("Select Duration for the Report", &telebot.SendOptions{
		ReplyTo: ctx.Message(),
		ReplyMarkup: &telebot.ReplyMarkup{
			InlineKeyboard: generateInlineKeyboard(inlineButtons),
			ForceReply:     true,
		},
	})
}

func generateReportDurationInlineButton(callbackOpts CallbackOptions) []telebot.InlineButton {
	durations := []SummaryDuration{DurationOneWeek, DurationThisMonth, DurationOneMonth, DurationHalfYear, DurationThisYear, DurationOneYear}
	inlineButtons := make([]telebot.InlineButton, 0, 3)
	for _, duration := range durations {
		callbackOpts.Report.Duration = duration
		btn := generateInlineButton(callbackOpts, duration)
		inlineButtons = append(inlineButtons, btn)
	}

	return inlineButtons
}

func handleReportCallback(ctx telebot.Context, callbackOpts CallbackOptions) error {
	report, err := generateReport(ctx, callbackOpts.Report)
	if err != nil {
		return ctx.Send(models.ErrCommonResponse(err))
	}

	//if err  = generateSampleJsonReport(report); err != nil {
	//	return ctx.Send(models.ErrCommonResponse(err))
	//}

	if err = generateTransactionReportFromTemplate(report, ""); err != nil {
		return ctx.Send(err.Error())
	}

	return ctx.Send(&telebot.Document{
		File:     telebot.FromDisk("/tmp/transaction_report.pdf"),
		FileName: "transaction_report.pdf",
	})
}

func generateSampleJsonReport(report gqtypes.Report) error { //nolint:unused // kept for local debugging
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pkg.ProjectDirectory+"/templates/"+"sample_report.json", data, 0644)
}

func generateReport(ctx telebot.Context, rop ReportCallbackOptions) (gqtypes.Report, error) {
	now, startTime := time.Now(), calculateStartTime(rop.Duration)

	svc := all.GetServices()
	user, err := svc.User.GetUserByTelegramID(ctx.Sender().ID)
	if err != nil {
		return gqtypes.Report{}, err
	}

	txns, err := svc.Txn.ListTransactionsByTime(user.ID, "", startTime.Unix(), now.Unix())
	if err != nil {
		return gqtypes.Report{}, err
	}

	report := gqtypes.Report{
		Name:      fmt.Sprintf("%v %v", user.FirstName, user.LastName),
		StartDate: startTime,
		EndDate:   now,
	}
	txnApis := make([]gqtypes.Transaction, 0, len(txns))
	for _, txn := range txns {
		txnApis = append(txnApis, convert.ToTransactionAPIFormat(txn))
	}

	report.Transactions = txnApis

	summary := gqtypes.SummaryGroups{
		Type:        map[string]gqtypes.FieldCost{},
		Category:    map[string]gqtypes.FieldCost{},
		Subcategory: map[string]gqtypes.FieldCost{},
	}
	for _, txn := range txns {
		// summarize transaction types
		fc := summary.Type[string(txn.Type)]
		fc.Amount += txn.Amount
		summary.Type[string(txn.Type)] = fc

		// summarize transaction subcategories
		fc = summary.Subcategory[txn.SubcategoryID]
		fc.Amount += txn.Amount
		summary.Subcategory[txn.SubcategoryID] = fc

		// summarize transaction categories
		cat := strings.Split(txn.SubcategoryID, "-")[0]
		fc = summary.Category[cat]
		fc.Amount += txn.Amount
		summary.Category[cat] = fc
	}

	for k, fc := range summary.Type {
		fc.Name = k
		summary.Type[k] = fc
	}

	for k, fc := range summary.Category {
		fc.Name, err = svc.Txn.GetTxnCategoryName(k)
		if err != nil {
			return gqtypes.Report{}, err
		}

		summary.Category[k] = fc
	}

	for k, fc := range summary.Subcategory {
		fc.Name, err = svc.Txn.GetTxnSubcategoryName(k)
		if err != nil {
			return gqtypes.Report{}, err
		}

		summary.Subcategory[k] = fc
	}

	report.Summary = summary
	return report, nil
}

func calculateStartTime(duration SummaryDuration) time.Time {
	now, startTime := time.Now(), pkg.StartOfMonth()
	switch duration {
	case DurationOneWeek:
		startTime = now.AddDate(0, 0, -7)
	case DurationThisMonth:
		startTime = pkg.StartOfMonth()
	case DurationOneMonth:
		startTime = now.AddDate(0, -1, 0)
	case DurationHalfYear:
		startTime = now.AddDate(0, -6, 0)
	case DurationThisYear:
		startTime = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	case DurationOneYear:
		startTime = now.AddDate(-1, 0, 0)
	}
	return startTime
}

func generateTransactionReportFromTemplate(report gqtypes.Report, reportPdfFileName string) error {
	// Process body template
	bodyData, err := templates.FS.ReadFile("transaction_report.tmpl")
	if err != nil {
		return err
	}
	bodyBuf := bytes.Buffer{}
	funcMap := template.FuncMap{"formatBDT": FormatBDT}
	bodyTmpl, err := template.New("report").Funcs(funcMap).Parse(string(bodyData))
	if err != nil {
		return err
	}
	if err = bodyTmpl.Execute(&bodyBuf, &report); err != nil {
		return err
	}
	if err = os.WriteFile("/tmp/transaction_report.html", bodyBuf.Bytes(), 0644); err != nil { //nolint:gosec // temp file for PDF pipeline
		return err
	}

	// Process header template
	headerData, err := templates.FS.ReadFile("header.tmpl")
	if err != nil {
		return err
	}
	headerBuf := bytes.Buffer{}
	headerTmpl, err := template.New("header").Funcs(funcMap).Parse(string(headerData))
	if err != nil {
		return err
	}
	if err = headerTmpl.Execute(&headerBuf, &report); err != nil {
		return err
	}
	if err = os.WriteFile("/tmp/header.html", headerBuf.Bytes(), 0644); err != nil { //nolint:gosec // temp file for PDF pipeline
		return err
	}

	// Read footer template
	footerData, err := templates.FS.ReadFile("footer.tmpl")
	if err != nil {
		return err
	}
	if err = os.WriteFile("/tmp/footer.html", footerData, 0644); err != nil { //nolint:gosec // temp file for PDF pipeline
		return err
	}

	if reportPdfFileName == "" {
		reportPdfFileName = "/tmp/transaction_report.pdf"
	}
	return pkg.ConvertHTMLToPDF(
		configs.TrackerConfig.System.PDFConverter,
		reportPdfFileName,
		bodyBuf.Bytes(),
		headerBuf.Bytes(),
		footerData,
	)
}

func FormatBDT(amount float64) string {
	if amount < 0 {
		return "-" + FormatBDT(-amount)
	}

	// Split into integer and decimal parts
	integerPart := int64(amount)
	decimalPart := amount - float64(integerPart)

	// Format integer part with BDT commas
	formattedInteger := formatBangladeshiCommas(integerPart)

	// Handle decimal part
	if decimalPart == 0 {
		return formattedInteger
	}

	// Format decimal to 2 places and trim leading zero
	decimalStr := strings.TrimPrefix(strconv.FormatFloat(decimalPart, 'f', 2, 64), "0")
	return formattedInteger + decimalStr
}

func formatBangladeshiCommas(n int64) string {
	str := strconv.FormatInt(n, 10)
	length := len(str)

	if length <= 3 {
		return str
	}

	// Reverse string for easier grouping
	reversed := reverseString(str)
	groups := []string{}

	// First group of 3 digits
	groups = append(groups, reversed[:3])
	reversed = reversed[3:]

	// Subsequent groups of 2 digits
	for len(reversed) > 0 {
		chunkSize := 2
		if len(reversed) < chunkSize {
			chunkSize = len(reversed)
		}
		groups = append(groups, reversed[:chunkSize])
		reversed = reversed[chunkSize:]
	}

	// Reverse groups and create formatted string
	formattedGroups := make([]string, len(groups))
	for i, group := range groups {
		formattedGroups[len(groups)-1-i] = reverseString(group)
	}

	return strings.Join(formattedGroups, ",")
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
