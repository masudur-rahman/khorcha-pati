package gqtypes

import (
	"bytes"
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

type Transaction struct {
	Date        time.Time `json:"date"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	Person      string    `json:"person"`
	Category    string    `json:"category"`
	Subcategory string    `json:"subcategory"`
	Remarks     string    `json:"remarks"`
}

type Summary struct {
	Income  float64
	Expense float64
}

func (s Summary) String() string {
	return fmt.Sprintf(`
Transaction Summary

Income:  %v
Expense: %v
`, s.Income, s.Income)
}

type FieldCost struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type SummaryGroups struct {
	Type        map[string]FieldCost `json:"type"`
	Category    map[string]FieldCost `json:"category"`
	Subcategory map[string]FieldCost `json:"subcategory"`
}

type Report struct {
	Name         string        `json:"name"`
	Transactions []Transaction `json:"transactions"`
	Summary      SummaryGroups `json:"summary"`
	StartDate    time.Time     `json:"startDate"`
	EndDate      time.Time     `json:"endDate"`
}

func (s SummaryGroups) String() string {
	buf := bytes.Buffer{}
	w := tabwriter.NewWriter(&buf, 0, 0, 5, ' ', 0)
	fmt.Fprintln(w, fmt.Sprintf("Transaction Summary\n"))

	for k, v := range s.Type {
		f := v.Name
		if f == "" {
			f = k
		}
		fmt.Fprintln(w, fmt.Sprintf("%v:\t%.2f", f, v.Amount))
	}

	for k, v := range s.Category {
		f := v.Name
		if f == "" {
			f = k
		}
		fmt.Fprintln(w, fmt.Sprintf("%v:\t%.2f", f, v.Amount))
	}

	for k, v := range s.Subcategory {
		f := v.Name
		if f == "" {
			f = k
		}
		fmt.Fprintln(w, fmt.Sprintf("%v:\t%.2f", f, v.Amount))
	}

	_ = w.Flush()
	return buf.String()
}

func (s SummaryGroups) MarkdownString() string {
	buf := bytes.Buffer{}
	w := tabwriter.NewWriter(&buf, 0, 0, 5, ' ', 0)
	fmt.Fprintln(w, "\n## Transaction Summary")

	if len(s.Type) > 0 {
		writeRowHeader(w, "Type", "Amount")
	}
	for k, v := range s.Type {
		f := v.Name
		if f == "" {
			f = k
		}
		fmt.Fprintln(w, fmt.Sprintf("| %v | %v |", f, v.Amount))
	}

	if len(s.Type) > 0 {
		writeRowHeader(w, "Category", "Amount")
	}
	for k, v := range s.Category {
		f := v.Name
		if f == "" {
			f = k
		}
		fmt.Fprintln(w, fmt.Sprintf("| %v | %v |", f, v.Amount))
	}

	if len(s.Type) > 0 {
		writeRowHeader(w, "Subcategory", "Amount")
	}
	for k, v := range s.Subcategory {
		f := v.Name
		if f == "" {
			f = k
		}
		fmt.Fprintln(w, fmt.Sprintf("| %v | %v |", f, v.Amount))
	}

	_ = w.Flush()
	return buf.String()
}

func writeRowHeader(w io.Writer, a, b any) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, fmt.Sprintf("| %v | %v |", a, b))
	fmt.Fprintln(w, fmt.Sprintf("| --- | --- |"))
}
