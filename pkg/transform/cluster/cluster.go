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
	Resources   []ReportResource                                `json:"unsupportedResources,omitempty"`
	SrcOnlyRGs  map[string]map[string][]schema.GroupVersionKind `json:"sourceOnlyResources,omitempty"`
	GapGVKs     map[string]map[string][]schema.GroupVersionKind `json:"gapGVKs,omitempty"`
}

// ReportResource represents json data of resources
type ReportResource struct {
	ResourceName  string                    `json:"resourceName"`
	NamespaceList []string                  `json:"namespaces,omitempty"`
	Source        []schema.GroupVersionKind `json:"sourceGVKs,omitempty"`
	Destination   []schema.GroupVersionKind `json:"destinationGVKs,omitempty"`
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
	SrcOnlyRGs  map[string]map[string][]schema.GroupVersionKind `json:"sourceOnlyResources,omitempty"`
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
	clusterReport.Resources = GenMigResourceReport(apiResources.ResourceList)
	clusterReport.GapGVKs = apiResources.SrcGapRGVKs
	clusterReport.SrcOnlyRGs = apiResources.SrcOnlyRGs
	return
}

// GenMigResourceReport inserts report values for Source Cluster for json output
func GenMigResourceReport(apiResources []api.Resource) (ResourcesReport []ReportResource) {
	ResourcesReport = []ReportResource{}
	for _, apiResource := range apiResources {
		resource := ReportResource{}
		resource.ResourceName = apiResource.ResourceName
		resource.NamespaceList = apiResource.NamespaceList
		resource.Source = apiResource.Source
		resource.Destination = apiResource.Destination

		ResourcesReport = append(ResourcesReport, resource)
	}

	return
}

// GenSrcClusterReport inserts report values for Source Cluster for json output
func GenSrcClusterReport(apiResources api.Resources) (clusterReport ReportCluster) {
	clusterReport.ClusterName = api.SrcClusterName
	clusterReport.SrcOnlyRGs = apiResources.SrcOnlyRGs
	clusterReport.GapGVKs = apiResources.SrcGapRGVKs
	clusterReport.GVRs = apiResources.SrcRGVKs
	return
}

// GenDstClusterReport inserts report values for Destination Cluster for json output
func GenDstClusterReport(apiResources api.Resources) (clusterReport ReportCluster) {
	clusterReport.ClusterName = api.DstClusterName
	// clusterReport.DstOnlyGVKs = apiResources.DstOnlyGVKs
	clusterReport.GapGVKs = apiResources.DstGapRGVKs
	clusterReport.GVRs = apiResources.DstRGVKs
	return
}
