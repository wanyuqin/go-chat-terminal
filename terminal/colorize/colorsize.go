package colorize

import (
	"github.com/fatih/color"
)

func FgHiCyan(s string) string {
	cp := color.New()
	cp.Add(color.FgHiCyan)
	return cp.Sprintf(s)
}

func FgHiGreen(s string) string {
	cp := color.New()
	cp.Add(color.FgHiGreen)
	return cp.Sprintf(s)
}
