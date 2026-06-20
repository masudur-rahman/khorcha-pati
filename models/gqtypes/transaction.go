package gqtypes

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"sort"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/models"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

type Transaction struct {
	Date           time.Time `json:"date"`
	Type           string    `json:"type"`
	Amount         float64   `json:"amount"`
	Source         string    `json:"source"`
	Destination    string    `json:"destination"`
	Person         string    `json:"person"`
	Category       string    `json:"category"`
	Subcategory    string    `json:"subcategory"`
	Remarks        string    `json:"remarks"`
	RunningBalance float64   `json:"runningBalance"`
}

type Summary struct {
	Income  float64
	Expense float64
}

func (s Summary) String() string {
	return fmt.Sprintf(`
Transaction Summary

Income:  %s
Expense: %s
`, models.FormatMoney(s.Income), models.FormatMoney(s.Expense))
}

type FieldCost struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Type   string  `json:"type,omitempty"`
}

type SummaryGroups struct {
	Type        map[string]FieldCost `json:"type"`
	Category    map[string]FieldCost `json:"category"`
	Subcategory map[string]FieldCost `json:"subcategory"`
}

type Wallet struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	ShortName string    `json:"shortName"`
	Name      string    `json:"name"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
}

type Contact struct {
	ID               int64     `json:"id"`
	NickName         string    `json:"nickName"`
	FullName         string    `json:"fullName"`
	Email            string    `json:"email"`
	NetBalance       float64   `json:"netBalance"`
	LastTxnTimestamp int64     `json:"lastTxnTimestamp"`
	CreatedAt        time.Time `json:"createdAt"`
}

type Report struct {
	Name         string        `json:"name"`
	Transactions []Transaction `json:"transactions"`
	Summary      SummaryGroups `json:"summary"`
	Wallets      []Wallet      `json:"wallets,omitempty"`
	Contacts     []Contact     `json:"contacts,omitempty"`
	StartDate    time.Time     `json:"startDate"`
	EndDate      time.Time     `json:"endDate"`

	// Pre-computed sorted slices for template rendering
	TypeSummary        []FieldCost `json:"typeSummary,omitempty"`
	CategorySummary    []FieldCost `json:"categorySummary,omitempty"`
	SubcategorySummary []FieldCost `json:"subcategorySummary,omitempty"`
	TotalAmount        float64     `json:"totalAmount"`
	NetBalance         float64     `json:"netBalance"`
	GeneratedAt        time.Time   `json:"generatedAt"`
}

// SortMapToSlice converts map to sorted slice (Amount Descending).
func SortMapToSlice(input map[string]FieldCost) []FieldCost {
	var sorted []FieldCost
	for k, v := range input {
		if v.Name == "" {
			v.Name = k
		}
		sorted = append(sorted, v)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Amount > sorted[j].Amount
	})
	return sorted
}

const amountColWidth = 10

func printSection(w io.Writer, title string, data map[string]FieldCost) {
	if len(data) == 0 {
		return
	}
	items := SortMapToSlice(data)

	maxNameLen := 0
	for _, item := range items {
		if len(item.Name) > maxNameLen {
			maxNameLen = len(item.Name)
		}
	}

	nameColWidth := maxNameLen + 2

	fmt.Fprintln(w, title)

	fmt.Fprintf(w, "%-*s %*s\n", nameColWidth, "----", amountColWidth, "-------")

	for _, item := range items {
		fmt.Fprintf(w, "%-*s %*.2f\n", nameColWidth, item.Name, amountColWidth, item.Amount)
	}
	fmt.Fprintln(w, "")
}

func (s SummaryGroups) String() string {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "Transaction Summary")

	printSection(&buf, "BY TYPE", s.Type)
	printSection(&buf, "BY CATEGORY", s.Category)
	printSection(&buf, "BY SUBCATEGORY", s.Subcategory)

	return buf.String()
}

// MarkdownString stays the same
func (s SummaryGroups) MarkdownString() string {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "Transaction Summary")

	printMD := func(title string, data map[string]FieldCost) {
		if len(data) == 0 {
			return
		}
		items := SortMapToSlice(data)
		fmt.Fprintf(&buf, "\n### %s\n", title)
		fmt.Fprintln(&buf, "| Name | Amount |")
		fmt.Fprintln(&buf, "| --- | ---: |")
		for _, item := range items {
			fmt.Fprintf(&buf, "| %s | %s |\n", item.Name, models.FormatMoney(item.Amount))
		}
	}

	printMD("Type", s.Type)
	printMD("Category", s.Category)
	printMD("Subcategory", s.Subcategory)

	return buf.String()
}

// --- Styling Constants ---
const (
	imgWidth       = 400
	margin         = 30.0
	rowHeight      = 35.0
	fontSizeTitle  = 24.0
	fontSizeHeader = 18.0
	fontSizeBody   = 18.0
)

// Define colors
var (
	bgCol       = color.RGBA{250, 250, 250, 255} // Off-white background
	textDarkCol = color.RGBA{50, 50, 50, 255}    // Dark grey text
	lineCol     = color.RGBA{200, 200, 200, 255} // Light grey for separators
	accentCol   = color.RGBA{0, 122, 255, 255}   // Blue for main titles
)

func (s SummaryGroups) PNG() ([]byte, error) {
	// 1. Load Font
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return nil, fmt.Errorf("could not load font: %w", err)
	}

	// 2. Calculate Dynamic Height
	totalRows := len(s.Type) + len(s.Category) + len(s.Subcategory)
	estimatedHeight := margin + 60 + (float64(totalRows)*rowHeight + 3*(rowHeight*3.5)) + margin

	// 3. Setup Graphics Context
	dc := gg.NewContext(imgWidth, int(estimatedHeight))
	dc.SetColor(bgCol)
	dc.Clear()

	setFontSize := func(size float64) {
		face := truetype.NewFace(font, &truetype.Options{Size: size, DPI: 72})
		dc.SetFontFace(face)
	}

	// --- Start Drawing ---
	currentY := margin + fontSizeTitle

	// Draw Main Report Title
	dc.SetColor(accentCol)
	setFontSize(fontSizeTitle)
	dc.DrawStringAnchored("Transaction Summary", imgWidth/2, currentY, 0.5, 0.5)
	currentY += 50

	// Helper to draw a specific section with CUSTOM HEADER
	drawSection := func(headerName string, data map[string]FieldCost) {
		if len(data) == 0 {
			return
		}
		items := SortMapToSlice(data)

		// Section Title (e.g., "BY CATEGORY")
		dc.SetColor(textDarkCol)
		setFontSize(fontSizeHeader)
		//dc.DrawString(title, margin, currentY)
		//currentY += rowHeight * 1.2

		// Header Row (Custom Name | Amount)
		setFontSize(fontSizeBody)
		dc.SetColor(textDarkCol)

		// Use the custom headerName here
		dc.DrawString(headerName, margin, currentY)

		dc.DrawStringAnchored("Amount", imgWidth-margin, currentY, 1, 0)
		currentY += 10

		// Draw Separator Line
		dc.SetColor(lineCol)
		dc.SetLineWidth(1)
		dc.DrawLine(margin, currentY, imgWidth-margin, currentY)
		dc.Stroke()
		currentY += rowHeight

		// Draw Data Rows
		setFontSize(fontSizeBody)
		dc.SetColor(textDarkCol)
		for _, item := range items {
			dc.DrawStringAnchored(item.Name, margin, currentY, 0, 0.5)

			amtStr := models.GroupAmount(item.Amount, models.DefaultGrouping)
			dc.DrawStringAnchored(amtStr, imgWidth-margin, currentY, 1, 0.5)

			currentY += rowHeight
		}
		currentY += rowHeight * 0.8
	}

	// Draw the three sections with distinct headers
	drawSection("Transaction Type", s.Type)
	drawSection("Category", s.Category)
	drawSection("Subcategory", s.Subcategory)

	// 4. Encode to PNG Buffer (Cropped to exact size)
	var buf bytes.Buffer
	// Check if interface conversion is needed (gg usually returns *image.RGBA)
	if rgbaImg, ok := dc.Image().(*image.RGBA); ok {
		finalImage := rgbaImg.SubImage(image.Rect(0, 0, imgWidth, int(currentY+margin)))
		err = png.Encode(&buf, finalImage)
	} else {
		// Fallback if image type differs
		err = png.Encode(&buf, dc.Image())
	}

	if err != nil {
		return nil, fmt.Errorf("could not encode png: %w", err)
	}

	return buf.Bytes(), nil
}
