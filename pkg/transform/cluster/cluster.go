package cluster

import (
	"github.com/gildub/phronetic/pkg/api"
	"github.com/sirupsen/logrus"

	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Report represents json report of k8s resources
type Report struct {
	Namespace   NamespaceReport                                 `json:"namespace,omitempty"`
	SrcOnlyGVKs map[string]map[string][]schema.GroupVersionKind `json:"sourceOnlyGroupVersionKinds,omitempty"`
	SrcGapGVKs  map[string]map[string][]schema.GroupVersionKind `json:"sourceGapGroupVersionKinds,omitempty"`
	DstGapGVKs  map[string]map[string][]schema.GroupVersionKind `json:"destinationGapGroupVersionKinds,omitempty"`
}

// NamespaceReport represents json report of k8s namespaces
type NamespaceReport struct {
	Name         string       `json:"name"`
	LatestChange k8sMeta.Time `json:"latestChange,omitempty"`
}

// GroupVersionsReport represents json report of k8s storage classes
type GroupVersionsReport struct {
	GroupVersion string `json:"groupversion"`
}

// GenClusterReport inserts report values into structures for json output
func GenClusterReport(apiResources api.Resources) (clusterReport Report) {
	clusterReport.ReportNamespaceResources(apiResources.NamespaceResources)
	clusterReport.SrcOnlyGVKs = apiResources.SrcOnlyGVKs
	clusterReport.SrcGapGVKs = apiResources.SrcGapGVKs
	clusterReport.DstGapGVKs = apiResources.DstGapGVKs
	return
}

// ReportNamespaceResources fills in information about resources of a namespace
func (clusterReport *Report) ReportNamespaceResources(apiResources *api.NamespaceResources) {
	logrus.Info("ClusterReport::ReportNamespaceResources")

	reportedNamespace := NamespaceReport{Name: apiResources.Namespace.Name}
	clusterReport.Namespace = reportedNamespace
}
