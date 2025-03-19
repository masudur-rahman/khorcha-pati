package pkg

import (
	"context"
	"net/url"
	"os"
	"os/exec"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func ConvertHTMLToPDF(converter string, outputFile string, data []byte) error {
	switch converter {
	case "wkhtmltopdf":
		return convertViaWkhtmlToPDF(outputFile, data)
	case "chromedp":
		return convertViaChromeDP(outputFile, data)
	default:
		return convertViaChromeDP(outputFile, data)

	}
}

func convertViaWkhtmlToPDF(outputFile string, data []byte) error {
	inputFile := "/tmp/document.html"
	if err := os.WriteFile(inputFile, data, 0644); err != nil {
		return err
	}

	return exec.Command("wkhtmltopdf", "--title", "Transaction Report",
		"--zoom", "0.98", "--page-size", "A4", "--orientation", "Portrait", inputFile, outputFile).Run()
}

func convertViaChromeDP(outputFile string, htmlContent []byte) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("font-render-hinting", "none"), // Better font rendering
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var pdfBuf []byte
	dataURL := "data:text/html," + url.PathEscape(string(htmlContent))

	err := chromedp.Run(ctx,
		chromedp.Navigate(dataURL),
		chromedp.WaitReady("table"), // Wait for tables to render
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			// PDF parameters combining best of both outputs
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(8.27).   // A4 width in inches (210mm)
				WithPaperHeight(11.69). // A4 height in inches (297mm)
				WithMarginTop(0.2).
				WithMarginBottom(0.2).
				WithMarginLeft(0.2).
				WithMarginRight(0.2).
				WithScale(0.80). // Compromise between zoom 0.96 and full size
				WithPreferCSSPageSize(true).
				Do(ctx)
			return err
		}),
	)

	if err != nil {
		return err
	}
	return os.WriteFile(outputFile, pdfBuf, 0644)
}
