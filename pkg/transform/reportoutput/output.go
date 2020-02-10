package reportoutput

import (
	"github.com/gildub/phronetic/pkg/transform/cluster"
)

// ReportOutput holds a collection of reports to be written to file
type ReportOutput struct {
	MigOperatorReport cluster.ReportMigOperator `json:"migOperator,omitempty"`
	DiffReport        cluster.ReportDiff        `json:"differential,omitempty"`
}

var (
	jsonFileName = "report.json"
)

// DumpReports creates OCDs files
func DumpReports(r ReportOutput) {
	jsonOutput(r)
}
