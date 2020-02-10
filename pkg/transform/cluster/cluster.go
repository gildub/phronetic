package cluster

import (
	"github.com/gildub/phronetic/pkg/api"
	"github.com/sirupsen/logrus"

	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Report represents json report of k8s resources
type Report struct {
	ClusterName string
	Namespace   NamespaceReport                                 `json:"namespace,omitempty"`
	GVRs        map[string]map[string][]schema.GroupVersionKind `json:"resourcesGroupVersionKinds,omitempty"`
	OnlyGVKs    map[string]map[string][]schema.GroupVersionKind `json:"onlyGroupVersionKinds,omitempty"`
	GapGVKs     map[string]map[string][]schema.GroupVersionKind `json:"gapGroupVersionKinds,omitempty"`
}

// NamespaceReport represents json report of k8s namespaces
type NamespaceReport struct {
	Name         string       `json:"name"`
	LatestChange k8sMeta.Time `json:"latestChange,omitempty"`
}

// ReportNamespaceResources fills in information about resources of a namespace
func (clusterReport *Report) ReportNamespaceResources(apiResources *api.NamespaceResources) {
	logrus.Info("ClusterReport::ReportNamespaceResources")

	reportedNamespace := NamespaceReport{Name: apiResources.Namespace.Name}
	clusterReport.Namespace = reportedNamespace
}

// GenSrcClusterReport inserts report values for Source Cluster for json output
func GenSrcClusterReport(apiResources api.Resources) (clusterReport Report) {
	clusterReport.ClusterName = api.SrcClusterName
	clusterReport.OnlyGVKs = apiResources.SrcOnlyRGVKs
	clusterReport.GapGVKs = apiResources.SrcGapRGVKs
	if api.CtrlClient != nil {
		clusterReport.GVRs = apiResources.SrcRGVKs
	}
	return
}

// GenDstClusterReport inserts report values for Destination Cluster for json output
func GenDstClusterReport(apiResources api.Resources) (clusterReport Report) {
	clusterReport.ClusterName = api.DstClusterName
	//clusterReport.DstOnlyGVKs = apiResources.DstOnlyGVKs
	clusterReport.GapGVKs = apiResources.DstGapRGVKs
	clusterReport.GVRs = apiResources.DstRGVKs
	return
}
