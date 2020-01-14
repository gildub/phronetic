package reportoutput

import (
	"github.com/gildub/phronetic/pkg/transform/cluster"
)

// ReportOutput holds a collection of reports to be written to file
type ReportOutput struct {
	ClusterReport cluster.Report `json:"cluster,omitempty"`
}

var (
	jsonFileName = "report.json"
)

// DumpReports creates OCDs files
func DumpReports(r ReportOutput) {
	jsonOutput(r)
}
