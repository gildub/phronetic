package transform

import (
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

// Transform converts the retrieved information to a useful output
func (e ClusterExtraction) Transform() ([]Output, error) {
	outputs := []Output{}
	logrus.Info("ClusterTransform::Transform:Reports")

	srcClusterReport := cluster.GenSrcClusterReport(e.Resources)
	FinalReportOutput.Report.SrcClusterReport = srcClusterReport

	if api.CtrlClient == nil {
		dstClusterReport := cluster.GenDstClusterReport(e.Resources)
		FinalReportOutput.Report.DstClusterReport = dstClusterReport
	}

	return outputs, nil
}

// Validate no need to validate it, data is exctracted from API
func (e ClusterExtraction) Validate() (err error) { return }

// Extract collects data for cluster report
func (e ClusterTransform) Extract() (Extraction, error) {
	extraction := &ClusterExtraction{}

	extraction.SrcOnlyRGVKs = map[string]map[string][]schema.GroupVersionKind{}
	extraction.SrcGapRGVKs = map[string]map[string][]schema.GroupVersionKind{}
	extraction.DstGapRGVKs = map[string]map[string][]schema.GroupVersionKind{}

	api.SrcRESTMapper = api.RESTMapperGetGRs(api.K8sSrcClient)
	api.DstRESTMapper = api.RESTMapperGetGRs(api.K8sDstClient)

	extraction.SrcRGVKs = listNamespacedResources(api.K8sSrcClient, api.SrcRESTMapper)
	extraction.DstRGVKs = listNamespacedResources(api.K8sDstClient, api.DstRESTMapper)

	namespace := env.Config().GetString("Namespace")
	namespaceResource := api.GetNamespace(api.K8sSrcClient, namespace)
	namespaceResources := api.NamespaceResources{Namespace: namespaceResource}
	extraction.NamespaceResources = &namespaceResources

	for srcRes, srcGroupGVKs := range extraction.SrcRGVKs {
		for srcGroup, srcGVs := range srcGroupGVKs {
			if dstGVs, ok := extraction.DstRGVKs[srcRes][srcGroup]; ok {
				if !sameGVKs(srcGVs, dstGVs) {
					if !commonGVKs(srcGVs, dstGVs) {
						if api.CtrlClient != nil {
						}
						curGVR := schema.GroupVersionResource{
							Group: srcGroup,
							// TODO: Replace with Preferred Version
							Version:  extraction.SrcRGVKs[srcRes][srcGroup][0].Version,
							Resource: srcRes,
						}
						crdClient := api.K8sSrcDynClient.Resource(curGVR)

						crd, err := crdClient.List(metav1.ListOptions{})
						if err != nil {
							logrus.Fatalf("Error getting CRD %v", err)
						}

						if crd != nil {
							extraction.SrcGapRGVKs[srcRes] = map[string][]schema.GroupVersionKind{}
							extraction.SrcGapRGVKs[srcRes][srcGroup] = srcGVs
							extraction.DstGapRGVKs[srcRes] = map[string][]schema.GroupVersionKind{}
							extraction.DstGapRGVKs[srcRes][srcGroup] = dstGVs
						}
					}
				}
			} else {
				extraction.SrcOnlyRGVKs[srcRes] = map[string][]schema.GroupVersionKind{}
				extraction.SrcOnlyRGVKs[srcRes][srcGroup] = srcGVs
			}
		}
	}

	return *extraction, nil

	// TODO: exception rule?
	// return nil, errors.New("Cluster Transform failed: Migration controller missing")
}

func commonGVKs(src, dst []schema.GroupVersionKind) bool {
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

// listNamespacedResources parses provided list of RESTMapper.APIGroupRessources
// then filters resources that are only namespaced
// and trims out resources with suffixes extensions (such as */status, */rollback, */scale etc. I.E deployments/status)
// and finaly returns GroupVersionKinds broken down by group for each resource.
func listNamespacedResources(client *kubernetes.Clientset, restMapper meta.RESTMapper) map[string]map[string][]schema.GroupVersionKind {
	//map[string][]schema.GroupVersionKind {
	resources := api.ListServerResources(client)
	list := make(map[string]map[string][]schema.GroupVersionKind)
	for _, resource := range resources {
		for _, APIResource := range resource.APIResources {
			if APIResource.Namespaced {
				name := APIResource.Name
				last := strings.LastIndex(APIResource.Name, "/")
				if last != -1 {
					name = APIResource.Name[0:last]
				}

				if _, ok := list[name]; !ok {
					list[name] = map[string][]schema.GroupVersionKind{}
					gvks := api.GetKindsFor(restMapper, name)

					for _, gvk := range gvks {
						// TODO: Handle the case of empty group which corresponds to legacy "core"
						if gvk.Group != "" {
							list[name][gvk.Group] = getGVsFrom(gvks, gvk.Group)
						}
					}
				}
			}
		}
	}
	return list
}

func getGVsFrom(GVs []schema.GroupVersionKind, group string) []schema.GroupVersionKind {
	list := []schema.GroupVersionKind{}
	for _, GV := range GVs {
		if GV.Group == group {
			list = append(list, GV)
		}
	}
	return list
}

// Name returns a human readable name for the transform
func (e ClusterTransform) Name() string {
	return ClusterTransformName
}
