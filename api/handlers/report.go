package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
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

	pdfFile, err := generateTransactionReportFromTemplate(report)
	if err != nil {
		return ctx.Send(err.Error())
	}
	defer os.Remove(pdfFile)

	return ctx.Send(&telebot.Document{
		File:     telebot.FromDisk(pdfFile),
		FileName: "transaction_report.pdf",
	})
}

func generateSampleJSONReport(report gqtypes.Report) error { //nolint:unused // kept for local debugging
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

	summary, err := buildSummary(svc, txns)
	if err != nil {
		return gqtypes.Report{}, err
	}
	report.Summary = summary

	report.TypeSummary = gqtypes.SortMapToSlice(summary.Type)
	report.CategorySummary, err = buildTypeSeparatedSummary(svc, txns, categoryKeyFn, svc.Txn.GetTxnCategoryName)
	if err != nil {
		return gqtypes.Report{}, err
	}
	report.SubcategorySummary, err = buildTypeSeparatedSummary(svc, txns, subcategoryKeyFn, svc.Txn.GetTxnSubcategoryName)
	if err != nil {
		return gqtypes.Report{}, err
	}

	report.TotalAmount, report.NetBalance = computeTotals(txns)

	return report, nil
}

func buildSummary(svc *all.Services, txns []models.Transaction) (gqtypes.SummaryGroups, error) {
	summary := gqtypes.SummaryGroups{
		Type:        map[string]gqtypes.FieldCost{},
		Category:    map[string]gqtypes.FieldCost{},
		Subcategory: map[string]gqtypes.FieldCost{},
	}

	for _, txn := range txns {
		fc := summary.Type[string(txn.Type)]
		fc.Amount += txn.Amount
		summary.Type[string(txn.Type)] = fc

		fc = summary.Subcategory[txn.SubcategoryID]
		fc.Amount += txn.Amount
		summary.Subcategory[txn.SubcategoryID] = fc

		cat := strings.Split(txn.SubcategoryID, "-")[0]
		fc = summary.Category[cat]
		fc.Amount += txn.Amount
		summary.Category[cat] = fc
	}

	for k, fc := range summary.Type {
		fc.Name = k
		summary.Type[k] = fc
	}

	var err error
	for k, fc := range summary.Category {
		fc.Name, err = svc.Txn.GetTxnCategoryName(k)
		if err != nil {
			return gqtypes.SummaryGroups{}, err
		}
		summary.Category[k] = fc
	}

	for k, fc := range summary.Subcategory {
		fc.Name, err = svc.Txn.GetTxnSubcategoryName(k)
		if err != nil {
			return gqtypes.SummaryGroups{}, err
		}
		summary.Subcategory[k] = fc
	}

	return summary, nil
}

func categoryKeyFn(txn models.Transaction) string {
	return strings.Split(txn.SubcategoryID, "-")[0]
}

func subcategoryKeyFn(txn models.Transaction) string {
	return txn.SubcategoryID
}

// buildTypeSeparatedSummary aggregates by key+type, resolves names, and returns sorted slice.
func buildTypeSeparatedSummary(
	svc *all.Services,
	txns []models.Transaction,
	keyFn func(models.Transaction) string,
	nameFn func(string) (string, error),
) ([]gqtypes.FieldCost, error) {
	type compositeKey struct {
		key     string
		txnType string
	}

	agg := map[compositeKey]float64{}
	for _, txn := range txns {
		k := compositeKey{key: keyFn(txn), txnType: string(txn.Type)}
		agg[k] += txn.Amount
	}

	result := make([]gqtypes.FieldCost, 0, len(agg))
	for ck, amount := range agg {
		name, err := nameFn(ck.key)
		if err != nil {
			return nil, err
		}
		result = append(result, gqtypes.FieldCost{
			Name:   name,
			Amount: amount,
			Type:   ck.txnType,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Amount > result[j].Amount
	})

	return result, nil
}

func computeTotals(txns []models.Transaction) (totalAmount, netBalance float64) {
	for _, txn := range txns {
		totalAmount += txn.Amount
		switch txn.Type {
		case models.IncomeTransaction:
			netBalance += txn.Amount
		case models.ExpenseTransaction:
			netBalance -= txn.Amount
		}
	}
	return totalAmount, netBalance
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

func generateTransactionReportFromTemplate(report gqtypes.Report) (string, error) {
	converter := configs.TrackerConfig.System.PDFConverter

	funcMap := template.FuncMap{
		"formatBDT": FormatBDT,
		"toLower":   strings.ToLower,
	}

	// Process body template
	bodyBytes, err := executeTemplate("transaction_report.tmpl", funcMap, &report)
	if err != nil {
		return "", err
	}

	// Select converter-specific header/footer templates
	headerTmplName, footerTmplName := selectTemplateNames(converter)

	headerBytes, err := executeTemplate(headerTmplName, funcMap, &report)
	if err != nil {
		return "", err
	}

	footerBytes, err := executeTemplate(footerTmplName, funcMap, &report)
	if err != nil {
		return "", err
	}

	pdfFile, err := os.CreateTemp("", "transaction_report_*.pdf")
	if err != nil {
		return "", err
	}
	pdfFile.Close()

	if err = pkg.ConvertHTMLToPDF(converter, pdfFile.Name(), bodyBytes, headerBytes, footerBytes); err != nil {
		os.Remove(pdfFile.Name())
		return "", err
	}

	return pdfFile.Name(), nil
}

// selectTemplateNames returns header and footer template names for the converter.
func selectTemplateNames(converter string) (header, footer string) {
	if converter == "chromedp" {
		return "header_cdp.tmpl", "footer_cdp.tmpl"
	}
	return "header.tmpl", "footer.tmpl"
}

// executeTemplate reads and executes a named template from the embedded FS.
func executeTemplate(name string, funcMap template.FuncMap, data any) ([]byte, error) {
	raw, err := templates.FS.ReadFile(name)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(name).Funcs(funcMap).Parse(string(raw))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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
