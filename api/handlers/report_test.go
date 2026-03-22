package handlers

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/masudur-rahman/expense-tracker-bot/configs"
	"github.com/masudur-rahman/expense-tracker-bot/models/gqtypes"
	"github.com/masudur-rahman/expense-tracker-bot/pkg"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTransactionInvoice_Wkhtmltopdf(t *testing.T) {
	if _, err := exec.LookPath("wkhtmltopdf"); err != nil {
		t.Skip("wkhtmltopdf not installed, skipping")
	}
	report, err := generateSampleReport()
	assert.NoError(t, err)
	configs.TrackerConfig.System.PDFConverter = "wkhtmltopdf"
	pdfPath, err := generateTransactionReportFromTemplate(report)
	assert.NoError(t, err)
	defer os.Remove(pdfPath)
	assert.FileExists(t, pdfPath)
}

func TestGenerateTransactionInvoice_Chromedp(t *testing.T) {
	report, err := generateSampleReport()
	assert.NoError(t, err)
	configs.TrackerConfig.System.PDFConverter = "chromedp"
	pdfPath, err := generateTransactionReportFromTemplate(report)
	assert.NoError(t, err)
	defer os.Remove(pdfPath)
	assert.FileExists(t, pdfPath)
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
