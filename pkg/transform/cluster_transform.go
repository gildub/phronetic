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
		SrcOnlyGVKs:        e.SrcOnlyGVKs,
		SrcGapGVKs:         e.SrcGapGVKs,
		DstGapGVKs:         e.DstGapGVKs,
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
		extraction.SrcOnlyGVKs = map[string]map[string][]schema.GroupVersionKind{}
		extraction.SrcGapGVKs = map[string]map[string][]schema.GroupVersionKind{}
		extraction.DstGapGVKs = map[string]map[string][]schema.GroupVersionKind{}

		namespace := env.Config().GetString("Namespace")
		namespaceResource := api.GetNamespace(api.K8sSrcClient, namespace)
		namespaceResources := api.NamespaceResources{Namespace: namespaceResource}
		extraction.NamespaceResources = &namespaceResources

		api.SrcRESTMapper = api.RESTMapperGetGRs(api.K8sSrcClient)
		api.DstRESTMapper = api.RESTMapperGetGRs(api.K8sDstClient)

		srcMap := listNamespacedResources(api.K8sSrcClient, api.SrcRESTMapper)
		dstMap := listNamespacedResources(api.K8sDstClient, api.DstRESTMapper)

		for srcRes, srcGroupGVKs := range srcMap {
			for srcGroup, srcGVs := range srcGroupGVKs {
				if dstGVs, ok := dstMap[srcRes][srcGroup]; ok {
					if !sameGVKs(srcGVs, dstGVs) {
						if !commonGVKs(srcGVs, dstGVs) {
							curGVR := schema.GroupVersionResource{
								Group: srcGroup,
								// TODO: Replace with Preferred Version
								Version:  srcMap[srcRes][srcGroup][0].Version,
								Resource: srcRes,
							}
							crdClient := api.K8sSrcDynClient.Resource(curGVR)

							crd, err := crdClient.List(metav1.ListOptions{})
							if err != nil {
								logrus.Fatalf("Error getting CRD %v", err)
							}

							if crd != nil {
								extraction.SrcGapGVKs[srcRes] = map[string][]schema.GroupVersionKind{}
								extraction.SrcGapGVKs[srcRes][srcGroup] = srcGVs
								extraction.DstGapGVKs[srcRes] = map[string][]schema.GroupVersionKind{}
								extraction.DstGapGVKs[srcRes][srcGroup] = dstGVs
							}
						}
					}
				} else {
					extraction.SrcOnlyGVKs[srcRes] = map[string][]schema.GroupVersionKind{}
					extraction.SrcOnlyGVKs[srcRes][srcGroup] = srcGVs
				}
			}
		}

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

// listNamespacedResources parses provided list of RESTMapper.APIGroupRessources
// then filters resources that are only namespaced
// and trims out resources with suffixes extensions (such as */status, */rollback, */scale etc. I.E deployments/status)
// and finaly returns GroupVersionKinds broken down by group for each resource.
// for example:
/*
[
	"cronjobs": [
		"batch": [
			{
				Group: "batch",
				Version: "v2alpha1",
				Kind: "CronJob",
			},
		],
	],
	"localsubjectaccessreviews": [
		"authorization.k8s.io": [
			{
				Group: "authorization.k8s.io",
				Version: "v1",
				Kind: "LocalSubjectAccessReview",
			},
			{
				Group: "authorization.k8s.io",
				Version: "v1beta1",
				Kind: "LocalSubjectAccessReview",
			},
		],
		"authorization.openshift.io": [
			{
				Group: "authorization.openshift.io",
				Version: "v1",
				Kind: "LocalSubjectAccessReview",
			},
		],
	],
	"scheduledjobs": [
		"batch": [
			{
				Group: "batch",
				Version: "v2alpha1",
				Kind: "ScheduledJob",
			},
		],
	],
]
*/
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
