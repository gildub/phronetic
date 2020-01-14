package transform

import (
	"errors"

	"github.com/gildub/phronetic/pkg/api"
	"github.com/gildub/phronetic/pkg/env"
	"github.com/gildub/phronetic/pkg/transform/cluster"
	"github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterTransformName is the cluster report name
const ClusterTransformName = "Cluster"

// ClusterExtraction holds data extracted from k8s API resources
type ClusterExtraction struct {
	api.Resources
}

// ClusterTransform reprents transform for k8s API resources
type ClusterTransform struct {
}

// Transform converts data collected from an OCP3 API into a useful output
func (e ClusterExtraction) Transform() ([]Output, error) {
	outputs := []Output{}
	logrus.Info("ClusterTransform::Transform:Reports")

	clusterReport := cluster.GenClusterReport(api.Resources{
		NamespaceResources: e.NamespaceResources,
		NewGroupVersions:   e.NewGroupVersions,
	})

	FinalReportOutput.Report.ClusterReport = clusterReport

	return outputs, nil
}

// Validate no need to validate it, data is exctracted from API
func (e ClusterExtraction) Validate() (err error) { return }

// Extract collects data for cluster report
func (e ClusterTransform) Extract() (Extraction, error) {
	if api.CtrlClient != nil {
		extraction := &ClusterExtraction{}
		namespace := env.Config().GetString("Namespace")
		namespaceResource := api.GetNamespace(api.K8sSrcClient, namespace)

		namespaceResources := api.NamespaceResources{Namespace: namespaceResource}
		namespaceResources.ResourceQuotaList = api.ListResourceQuotas(api.K8sSrcClient, namespaceResource.Name)
		namespaceResources.PodList = api.ListPods(api.K8sSrcClient, namespaceResource.Name)
		namespaceResources.DeploymentList = api.ListDeployments(api.K8sSrcClient, namespaceResource.Name)
		namespaceResources.DaemonSetList = api.ListDaemonSets(api.K8sSrcClient, namespaceResource.Name)
		namespaceResources.PVCList = api.ListPVCs(api.K8sSrcClient, namespaceResource.Name)

		extraction.NamespaceResources = &namespaceResources
		extraction.SrcGroupVersions = api.ListGroupVersions(api.K8sSrcClient)
		extraction.DstGroupVersions = api.ListGroupVersions(api.K8sDstClient)
		extraction.NewGroupVersions = NewGroupVersions(extraction.DstGroupVersions, extraction.DstGroupVersions)

		return *extraction, nil
	}

	return nil, errors.New("Cluster Transform failed: Migration controller missing")
}

// NewGroupVersions returns the list of new GroupVersions available in destination but in source
func NewGroupVersions(src *metav1.APIGroupList, dst *metav1.APIGroupList) []string {
	list := []string{}
	for _, dstGV := range filterGVs(dst) {
		found := false
		for _, srcGV := range filterGVs(src) {
			if dstGV == srcGV {
				found = true
			}
		}
		if found == false {
			list = append(list, dstGV)
		}
	}
	return list
}

// Name returns a human readable name for the transform
func (e ClusterTransform) Name() string {
	return ClusterTransformName
}

func filterGVs(gvs *metav1.APIGroupList) []string {
	list := []string{}
	for _, group := range gvs.Groups {
		for _, version := range group.Versions {
			list = append(list, version.GroupVersion)
		}
	}
	return list
}
