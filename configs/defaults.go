package configs

import (
	"sync"

	"github.com/masudur-rahman/khorcha-pati/pkg"

	"github.com/jedib0t/go-pretty/v6/table"
)

var (
	once    sync.Once
	printer pkg.Printer
)

func GetDefaultPrinter() pkg.Printer {
	once.Do(func() {
		opts := pkg.Options{Style: table.StyleColoredBright, EnableStdout: true}
		printer = pkg.NewPrinter(opts)
	})
	return printer
}
