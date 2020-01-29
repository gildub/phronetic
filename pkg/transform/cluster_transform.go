package transform

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gildub/phronetic/pkg/api"
	"github.com/gildub/phronetic/pkg/env"
	"github.com/gildub/phronetic/pkg/transform/cluster"
	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
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

		api.SrcRESTMapper = api.RESTMapperGetGRs(api.K8sSrcClient)
		api.DstRESTMapper = api.RESTMapperGetGRs(api.K8sDstClient)

		srcMap := namespacedResources(api.K8sSrcClient, api.SrcRESTMapper)
		dstMap := namespacedResources(api.K8sDstClient, api.DstRESTMapper)

		for srcRes, srcGVs := range srcMap {
			if dstGVs, ok := dstMap[srcRes]; ok {
				if !sameGVKs(srcGVs, dstGVs) {
					if !leastCommonGVKs(srcGVs, dstGVs) {
						fmt.Printf("CAN'T PORT: %q -> Source: %+v, Destination: %+v\n", srcRes, srcGVs, dstGVs)
					}
				}
			} else {
				fmt.Printf("SRC ONLY: %q => %+v\n", srcRes, srcGVs)
			}
		}

		namespace := env.Config().GetString("Namespace")
		namespaceResource := api.GetNamespace(api.K8sSrcClient, namespace)

		namespaceResources := api.NamespaceResources{Namespace: namespaceResource}
		namespaceResources.ResourceQuotaList = api.ListResourceQuotas(api.K8sSrcClient, namespaceResource.Name)
		namespaceResources.PodList = api.ListPods(api.K8sSrcClient, namespaceResource.Name)
		namespaceResources.DeploymentList = api.ListDeployments(api.K8sSrcClient, namespaceResource.Name)
		namespaceResources.DaemonSetList = api.ListDaemonSets(api.K8sSrcClient, namespaceResource.Name)
		namespaceResources.PVCList = api.ListPVCs(api.K8sSrcClient, namespaceResource.Name)
		namespaceResources.HPAList = api.ListHPAv1(api.K8sSrcClient, namespaceResource.Name)

		extraction.NamespaceResources = &namespaceResources

		return *extraction, nil
	}

	return nil, errors.New("Cluster Transform failed: Migration controller missing")
}

func leastCommonGVKs(src, dst []schema.GroupVersionKind) bool {
	for _, s := range src {
		for _, d := range dst {
			if s == d {
				return true
			}
		}
	}
	return false
}

func sameGVKs(src, dst []schema.GroupVersionKind) bool {
	if len(src) != len(dst) {
		return false
	}
	for _, s := range src {
		found := false
		for _, d := range dst {
			if s == d {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func namespacedResources(client *kubernetes.Clientset, restMapper meta.RESTMapper) map[string][]schema.GroupVersionKind {
	resources := api.ListServerResources(client)
	list := make(map[string][]schema.GroupVersionKind)
	for _, resource := range resources {
		for _, APIResource := range resource.APIResources {
			if APIResource.Namespaced {
				name := APIResource.Name
				last := strings.LastIndex(APIResource.Name, "/")
				if last != -1 {
					name = APIResource.Name[0:last]
				}
				if _, ok := list[name]; !ok {
					list[name] = api.GetKindsFor(restMapper, name)
				}
			}
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
