package transform

import (
	"errors"
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
					if !commonGVKs(srcGVs, dstGVs) {
						curGVR := schema.GroupVersionResource{
							Group:    srcMap[srcRes][0].Group,
							Version:  srcMap[srcRes][0].Version,
							Resource: srcRes,
						}
						crdClient := api.K8sSrcDynClient.Resource(curGVR)

						crd, err := crdClient.List(metav1.ListOptions{})
						if err != nil {
							logrus.Fatalf("Error getting CRD %v", err)
						}

						if crd != nil {

							logrus.Warningf("Source resource %q is incompatible for destination: Source: %+v, Destination: %+v\n", srcRes, srcGVs, dstGVs)
						}
					}
				}
			} else {
				logrus.Warningf("Source only resource %q => %+v\n", srcRes, srcGVs)
			}
		}

		namespace := env.Config().GetString("Namespace")
		namespaceResource := api.GetNamespace(api.K8sSrcClient, namespace)

		namespaceResources := api.NamespaceResources{Namespace: namespaceResource}
		extraction.NamespaceResources = &namespaceResources

		return *extraction, nil
	}

	return nil, errors.New("Cluster Transform failed: Migration controller missing")
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
