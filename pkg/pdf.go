package pkg

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ConvertHTMLToPDF converts HTML content to PDF using the specified generator.
func ConvertHTMLToPDF(generator string, outputFile string, data, header, footer []byte, title string) error {
	if title == "" {
		title = "Financial Transaction Statement"
	}
	switch generator {
	case "chromedp":
		return convertViaChromeDP(outputFile, data, header, footer)
	default:
		return convertViaWkhtmlToPDF(outputFile, data, header, footer, title)
	}
}

func convertViaWkhtmlToPDF(outputFile string, data, header, footer []byte, title string) error {
	inputFile, cleanInput, err := writeTempFile("wk-body-*.html", data)
	if err != nil {
		return err
	}
	defer cleanInput()

	headerFile, cleanHeader, err := writeTempFile("wk-header-*.html", header)
	if err != nil {
		return err
	}
	defer cleanHeader()

	footerFile, cleanFooter, err := writeTempFile("wk-footer-*.html", footer)
	if err != nil {
		return err
	}
	defer cleanFooter()

	return exec.Command("wkhtmltopdf",
		"--enable-local-file-access",
		"--encoding", "UTF-8",
		"--title", title,
		"--header-html", headerFile,
		"--footer-html", footerFile,
		"--margin-top", "30mm",
		"--header-spacing", "5",
		"--margin-bottom", "15mm",
		"--footer-spacing", "5",
		"--page-size", "A4",
		"--orientation", "Portrait",
		inputFile,
		outputFile).Run()
}

func convertViaChromeDP(outputFile string, htmlContent, header, footer []byte) error {
	tmpFile, cleanTmp, err := writeTempFile("cdp-body-*.html", htmlContent)
	if err != nil {
		return err
	}
	defer cleanTmp()

	fileURL := "file://" + tmpFile

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("font-render-hinting", "none"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var pdfBuf []byte

	err = chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			pdfBuf, _, err = page.PrintToPDF().
				WithDisplayHeaderFooter(true).
				WithHeaderTemplate(string(header)).
				WithFooterTemplate(string(footer)).
				WithPrintBackground(true).
				WithPaperWidth(8.27).
				WithPaperHeight(11.69).
				WithMarginTop(1.15).
				WithMarginBottom(0.5).
				WithMarginLeft(0.4).
				WithMarginRight(0.4).
				WithScale(0.78).
				Do(ctx)
			return err
		}),
	)

	if err != nil {
		return err
	}
	return os.WriteFile(outputFile, pdfBuf, 0644) //nolint:gosec // temp output file
}

// writeTempFile creates a temp file with the given content and returns its path and cleanup func.
func writeTempFile(pattern string, content []byte) (string, func(), error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", nil, fmt.Errorf("creating temp file: %w", err)
	}

	if _, err = f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", nil, fmt.Errorf("writing temp file: %w", err)
	}

	f.Close()
	return f.Name(), func() { os.Remove(f.Name()) }, nil
}
