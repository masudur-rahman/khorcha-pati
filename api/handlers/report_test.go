package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/masudur-rahman/khorcha-pati/configs"
	"github.com/masudur-rahman/khorcha-pati/models/gqtypes"
	"github.com/masudur-rahman/khorcha-pati/pkg"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTransactionInvoice_Wkhtmltopdf(t *testing.T) {
	if _, err := exec.LookPath("wkhtmltopdf"); err != nil {
		t.Skip("wkhtmltopdf not installed, skipping")
	}
	report, err := generateSampleReport()
	assert.NoError(t, err)
	configs.TrackerConfig.System.PDFGenerator = configs.PDFGeneratorWkhtmltopdf
	pdfPath, err := GenerateTransactionStatementFromTemplate(report, "")
	assert.NoError(t, err)
	//defer os.Remove(pdfPath)
	fmt.Println(pdfPath)
	assert.FileExists(t, pdfPath)
}

func TestGenerateTransactionInvoice_ChromeDP(t *testing.T) {
	report, err := generateSampleReport()
	assert.NoError(t, err)
	configs.TrackerConfig.System.PDFGenerator = configs.PDFGeneratorChromeDP
	pdfPath, err := GenerateTransactionStatementFromTemplate(report, "")
	assert.NoError(t, err)
	//defer os.Remove(pdfPath)
	fmt.Println(pdfPath)
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
