package pkg

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Printer interface {
	PrintDocument(doc interface{}) string
	PrintDocuments(docs interface{}) string

	WithColumns(columns ...string) Printer
	WithExceptColumns(exceptColumns []string) Printer
	WithAllColumns() Printer
	WithStyle(style table.Style) Printer
	WithStdout(stdout bool) Printer
	WithRenderType(typ RenderType) Printer

	ClearColumns()
}

type RenderType string

const (
	RenderTypeDefault  RenderType = "default"
	RenderTypeCSV      RenderType = "csv"
	RenderTypeHTML     RenderType = "html"
	RenderTypeMarkdown RenderType = "markdown"
)

type Options struct {
	Style        table.Style
	EnableStdout bool
	RenderType   RenderType
}

type printer struct {
	options Options
	columns []string
	xCols   []string
	allCols bool
}

func NewPrinter(options Options) Printer {
	return &printer{
		options: options,
	}
}

func (p *printer) PrintDocument(doc interface{}) string {
	t := p.newTableWriter()

	hr := p.generateHeaderRow(doc)
	dr := p.generateDataRow(doc)

	t.AppendHeader(hr)
	t.AppendRow(dr)

	return p.render(t)
}

func (p *printer) PrintDocuments(docs interface{}) string {
	t := p.newTableWriter()

	val := reflect.ValueOf(docs)
	if val.Kind() != reflect.Slice || val.Len() == 0 {
		return ""
	}

	doc := val.Index(0).Interface()

	hr := p.generateHeaderRow(doc)
	drs := p.generateDataRows(docs)

	t.AppendHeader(hr)
	for _, dr := range drs {
		t.AppendRow(dr)
	}

	table := fmt.Sprintf(`## %v
`, reflect.TypeOf(doc).Name())
	return table + p.render(t)
}

func (p *printer) WithColumns(columns ...string) Printer {
	p.columns = columns
	p.allCols = false
	return p
}

func (p *printer) WithExceptColumns(exceptColumns []string) Printer {
	p.xCols = exceptColumns
	p.allCols = false
	return p
}

func (p *printer) WithAllColumns() Printer {
	p.allCols = true
	return p
}

func (p *printer) WithStyle(style table.Style) Printer {
	p.options.Style = style
	return p
}

func (p *printer) WithStdout(stdout bool) Printer {
	p.options.EnableStdout = stdout
	return p
}

func (p *printer) WithRenderType(typ RenderType) Printer {
	p.options.RenderType = typ
	return p
}

func (p *printer) ClearColumns() {
	p.columns = nil
	p.xCols = nil
	p.allCols = false
}

func (p *printer) newTableWriter() table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(p.getWriter())
	t.SetStyle(p.options.Style)
	return t
}

func (p *printer) getWriter() io.Writer {
	if p.options.EnableStdout {
		return os.Stdout
	}

	return nil
}

func (p *printer) retrieveColumns(val reflect.Value) []string {
	if p.allCols || len(p.columns) == 0 {
		return p.retrieveAllColumns(val)
	}
	return p.columns
}

func (p *printer) retrieveAllColumns(val reflect.Value) []string {
	// Retrieve all field names from the struct type
	numFields := val.NumField()
	columns := make([]string, numFields)
	for idx := 0; idx < numFields; idx++ {
		columns[idx] = val.Type().Field(idx).Name
	}
	return columns
}

func (p *printer) generateHeaderRow(doc interface{}) table.Row {
	hr := table.Row{}
	val := reflect.ValueOf(doc)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	columns := p.retrieveColumns(val)
	for _, column := range columns {
		if !containsString(p.xCols, column) {
			hr = append(hr, column)
		}
	}

	return hr
}

func (p *printer) generateDataRows(docs interface{}) []table.Row {
	drs := []table.Row{}

	val := reflect.ValueOf(docs)
	if val.Kind() != reflect.Slice || val.Len() == 0 {
		return drs
	}

	for idx := 0; idx < val.Len(); idx++ {
		doc := val.Index(idx).Interface()
		dr := p.generateDataRow(doc)
		drs = append(drs, dr)
	}

	return drs
}

func (p *printer) generateDataRow(doc interface{}) table.Row {
	val := reflect.ValueOf(doc)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	dr := table.Row{}
	columns := p.retrieveColumns(val)
	for _, column := range columns {
		if containsString(p.xCols, column) {
			continue
		}
		field := val.FieldByName(column)
		if field.IsValid() {
			dr = append(dr, formatValue(field.Interface()))
		}
	}
	return dr
}

func (p *printer) render(t table.Writer) string {
	switch p.options.RenderType {
	case RenderTypeHTML:
		return t.RenderHTML()
	case RenderTypeCSV:
		return t.RenderCSV()
	case RenderTypeMarkdown:
		return t.RenderMarkdown()
	default:
		return t.Render()
	}
}

func formatValue(value interface{}) any {
	if reflect.TypeOf(value) == reflect.TypeOf(time.Time{}) {
		if t, ok := value.(time.Time); ok {
			return t.Format("Jan 02, 03:04 AM")
		}
	}
	return value
}

func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
