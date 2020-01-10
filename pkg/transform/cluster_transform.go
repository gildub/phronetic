package transform

import (
	"errors"

	"github.com/gildub/phronetic/pkg/api"
	"github.com/gildub/phronetic/pkg/transform/cluster"
	o7tapiauth "github.com/openshift/api/authorization/v1"
	o7tapiroute "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"

	"k8s.io/api/apps/v1beta1"
	k8sapicore "k8s.io/api/core/v1"
	extv1b1 "k8s.io/api/extensions/v1beta1"
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
		NamespaceList: e.NamespaceList,
		NewGVs:        e.NewGVs,
	})

	FinalReportOutput.Report.ClusterReport = clusterReport

	return outputs, nil
}

// Validate no need to validate it, data is exctracted from API
func (e ClusterExtraction) Validate() (err error) { return }

// Extract collects data for cluster report
func (e ClusterTransform) Extract() (Extraction, error) {
	if api.CtrlClient != nil {
		chanDstGVs := make(chan *metav1.APIGroupList)
		chanGVs := make(chan *metav1.APIGroupList)
		chanNamespaces := make(chan *k8sapicore.NamespaceList)

		if api.K8sDstClient != nil {
			go api.ListGroupVersions(api.K8sDstClient, chanDstGVs)
		}

		go api.ListGroupVersions(api.K8sClient, chanGVs)
		go api.ListNamespaces(api.K8sClient, chanNamespaces)
		extraction := &ClusterExtraction{}

		// Map all namespaces to their resources
		namespacesList := <-chanNamespaces
		namespaceListSize := len(namespacesList.Items)
		extraction.NamespaceList = make([]api.NamespaceResources, namespaceListSize, namespaceListSize)
		for i, namespace := range namespacesList.Items {
			namespaceResources := api.NamespaceResources{NamespaceName: namespace.Name}

			chanQuotas := make(chan *k8sapicore.ResourceQuotaList)
			chanPods := make(chan *k8sapicore.PodList)
			chanRoutes := make(chan *o7tapiroute.RouteList)
			chanDeployments := make(chan *v1beta1.DeploymentList)
			chanDaemonSets := make(chan *extv1b1.DaemonSetList)
			chanRoles := make(chan *o7tapiauth.RoleList)
			chanPVCs := make(chan *k8sapicore.PersistentVolumeClaimList)

			go api.ListResourceQuotas(api.K8sClient, namespace.Name, chanQuotas)
			go api.ListPods(api.K8sClient, namespace.Name, chanPods)
			go api.ListRoutes(api.O7tClient, namespace.Name, chanRoutes)
			go api.ListDeployments(api.K8sClient, namespace.Name, chanDeployments)
			go api.ListDaemonSets(api.K8sClient, namespace.Name, chanDaemonSets)
			go api.ListRoles(api.O7tClient, namespace.Name, chanRoles)
			go api.ListPVCs(api.K8sClient, namespace.Name, chanPVCs)

			namespaceResources.ResourceQuotaList = <-chanQuotas
			namespaceResources.PodList = <-chanPods
			namespaceResources.RouteList = <-chanRoutes
			namespaceResources.DeploymentList = <-chanDeployments
			namespaceResources.DaemonSetList = <-chanDaemonSets
			namespaceResources.RolesList = <-chanRoles
			namespaceResources.PVCList = <-chanPVCs

			extraction.NamespaceList[i] = namespaceResources
		}

		extraction.GroupVersions = <-chanGVs

		if api.K8sDstClient != nil {
			extraction.DstGroupVersions = <-chanDstGVs
			extraction.NewGVs = NewGroupVersions(extraction.GroupVersions, extraction.DstGroupVersions)
		}

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
