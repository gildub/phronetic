package cluster

import (
	"github.com/gildub/phronetic/pkg/api"
	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ReportMigOperator represents json report of CAM Operator results
type ReportMigOperator struct {
	ClusterName string                                          `json:"clusterName,omitempty"`
	Namespace   string                                          `json:"namespace,omitempty"`
	OnlyGVKs    map[string]map[string][]schema.GroupVersionKind `json:"onlyGroupVersionKinds,omitempty"`
	GapGVKs     map[string]map[string][]schema.GroupVersionKind `json:"gapGroupVersionKinds,omitempty"`
}

// ReportDiff represents json report of Cluster Differential report
type ReportDiff struct {
	ReportSrcCluster ReportCluster `json:"sourceCluster,omitempty"`
	ReportDstCluster ReportCluster `json:"destinationCluster,omitempty"`
}

// ReportCluster represents json report of Cluster Differential report
type ReportCluster struct {
	ClusterName string                                          `json:"clusterName,omitempty"`
	GVRs        map[string]map[string][]schema.GroupVersionKind `json:"resourcesGroupVersionKinds,omitempty"`
	OnlyGVKs    map[string]map[string][]schema.GroupVersionKind `json:"onlyGroupVersionKinds,omitempty"`
	GapGVKs     map[string]map[string][]schema.GroupVersionKind `json:"gapGroupVersionKinds,omitempty"`
}

// GenDiffReport inserts report values for Source Cluster for json output
func GenDiffReport(apiResources api.Resources) (clusterReport ReportDiff) {
	logrus.Info("ClusterReport::Report:Differential")
	clusterReport.ReportSrcCluster = GenSrcClusterReport(apiResources)
	clusterReport.ReportDstCluster = GenDstClusterReport(apiResources)
	return
}

// GenMigOperatorReport inserts report values for Source Cluster for json output
func GenMigOperatorReport(apiResources api.Resources) (clusterReport ReportMigOperator) {
	logrus.Info("ClusterReport::Report:MigOperator")
	clusterReport.ClusterName = api.SrcClusterName
	clusterReport.Namespace = apiResources.NamespaceResources.Namespace.Name
	clusterReport.GapGVKs = apiResources.SrcGapRGVKs
	clusterReport.OnlyGVKs = apiResources.SrcOnlyRGVKs
	return
}

// GenSrcClusterReport inserts report values for Source Cluster for json output
func GenSrcClusterReport(apiResources api.Resources) (clusterReport ReportCluster) {
	clusterReport.ClusterName = api.SrcClusterName
	clusterReport.OnlyGVKs = apiResources.SrcOnlyRGVKs
	clusterReport.GapGVKs = apiResources.SrcGapRGVKs
	clusterReport.GVRs = apiResources.SrcRGVKs
	return
}

// GenDstClusterReport inserts report values for Destination Cluster for json output
func GenDstClusterReport(apiResources api.Resources) (clusterReport ReportCluster) {
	clusterReport.ClusterName = api.DstClusterName
	//clusterReport.DstOnlyGVKs = apiResources.DstOnlyGVKs
	clusterReport.GapGVKs = apiResources.DstGapRGVKs
	clusterReport.GVRs = apiResources.DstRGVKs
	return
}
