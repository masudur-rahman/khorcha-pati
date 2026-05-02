package pkg

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/tabwriter"
	"time"
)

func FormatDocuments(docs any, cols ...string) string {
	docsElem := reflect.ValueOf(docs)
	if docsElem.Kind() != reflect.Slice || docsElem.Len() == 0 {
		return "No data found"
	}

	buf := bytes.Buffer{}
	w := tabwriter.NewWriter(&buf, 0, 0, 5, ' ', 0)
	fmt.Fprintln(w, strings.Join(cols, "\t"))

	for idx := 0; idx < docsElem.Len(); idx++ {
		doc := docsElem.Index(idx).Interface()
		var vals []string
		for _, col := range cols {
			elem := reflect.ValueOf(doc)
			if elem.Kind() == reflect.Pointer {
				elem = elem.Elem()
			}

			val := elem.FieldByName(col)
			if strings.Contains(strings.ToLower(col), "time") {
				switch val.Kind() {
				case reflect.Int64:
					vals = append(vals, time.Unix(val.Int(), 0).Format("Jan 02, 15:04"))
				default:
					vals = append(vals, val.MethodByName("Format").Call([]reflect.Value{reflect.ValueOf("Jan 02, 15:04")})[0].String())
				}
			} else {
				vals = append(vals, fmt.Sprintf("%v", val.Interface()))
			}
		}

		fmt.Fprintln(w, strings.Join(vals, "\t"))
	}
	_ = w.Flush()
	return buf.String()
}
