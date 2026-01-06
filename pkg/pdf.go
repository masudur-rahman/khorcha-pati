package pkg

import (
	"context"
	"net/url"
	"os"
	"os/exec"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func ConvertHTMLToPDF(converter string, outputFile string, data []byte, header, footer []byte) error {
	//return convertViaWkhtmlToPDF(outputFile, data)
	switch converter {
	case "wkhtmltopdf":
		return convertViaWkhtmlToPDF(outputFile, data, header, footer)
	case "chromedp":
		return convertViaChromeDP(outputFile, data, header, footer)
	default:
		return convertViaWkhtmlToPDF(outputFile, data, header, footer)
	}
}

func convertViaWkhtmlToPDF(outputFile string, data []byte, header, footer []byte) error {
	inputFile := "/tmp/document.html"
	if err := os.WriteFile(inputFile, data, 0644); err != nil {
		return err
	}

	headerFile := "/tmp/header.html"
	if err := os.WriteFile(headerFile, header, 0644); err != nil {
		return err
	}

	footerFile := "/tmp/footer.html"
	if err := os.WriteFile(footerFile, footer, 0644); err != nil {
		return err
	}

	return exec.Command("wkhtmltopdf",
		"--enable-local-file-access",
		"--encoding", "UTF-8",
		"--title", "Transaction Report",
		"--header-html", headerFile,
		"--footer-html", footerFile,
		"--margin-top", "30mm",
		"--header-spacing", "5", // Space between header and content
		"--margin-bottom", "15mm", // Bottom margin for footer
		"--footer-spacing", "5", // Space between footer and content
		"--page-size", "A4",
		"--orientation", "Portrait",
		inputFile,
		outputFile).Run()
}

func convertViaChromeDP(outputFile string, htmlContent []byte, header, footer []byte) error {
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
		chromedp.WaitReady("body"), // Wait for body to render
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			// PDF parameters combining best of both outputs
			pdfBuf, _, err = page.PrintToPDF().
				WithDisplayHeaderFooter(true).
				WithHeaderTemplate(string(header)).
				WithFooterTemplate(string(footer)).
				WithPrintBackground(true).
				WithPaperWidth(8.27).   // A4 width in inches (210mm)
				WithPaperHeight(11.69). // A4 height in inches (297mm)
				WithMarginTop(1.2).
				WithMarginBottom(0.5).
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
