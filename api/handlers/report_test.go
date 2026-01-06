package handlers

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/masudur-rahman/expense-tracker-bot/configs"
	"github.com/masudur-rahman/expense-tracker-bot/models/gqtypes"
	"github.com/masudur-rahman/expense-tracker-bot/pkg"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTransactionInvoice(t *testing.T) {
	report, err := generateSampleReport()
	assert.NoError(t, err)
	configs.TrackerConfig.System.PDFConverter = "wkhtmltopdf"
	//configs.TrackerConfig.System.PDFConverter = "chromedp"
	err = generateTransactionReportFromTemplate(report, "/tmp/transaction_report_test.pdf")
	assert.NoError(t, err)
}

func generateSampleReport() (gqtypes.Report, error) {
	data, err := os.ReadFile(pkg.ProjectDirectory + "/templates/" + "sample_report.json")
	if err != nil {
		return gqtypes.Report{}, err
	}

	report := gqtypes.Report{}
	if err = json.Unmarshal(data, &report); err != nil {
		return gqtypes.Report{}, err
	}
	return report, nil
}
