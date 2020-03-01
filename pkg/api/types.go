package api

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

/*
RGVKs is short for Resource -> GroupVersionKind
RGVKs represents resources broken down by group and their containing GVKs
Eaach list is built up and filtered down from k8s/client-go/restmapper.GetAPIGroupResources(client.Discovery())
A map[string]map[string][]schema.GroupVersionKind For example:
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
type Resources struct {
	ResourceList []Resource
	// SrcRGVKs contains all RGVKs available on source api-server (trimmed of "/.*"" suffixes)
	SrcRGVKs map[string]map[string][]schema.GroupVersionKind
	// DstRGVKs contains all RGVKs available on destination api-server (trimmed of "/.* suffixes)
	DstRGVKs map[string]map[string][]schema.GroupVersionKind
	// SrcOnlyRGs contains resources and the API Group only available on source api-server
	SrcOnlyRGs map[string]map[string][]schema.GroupVersionKind
	// SrcGapRGVKs contains RGVKs where group is in both source and destination api-servers but version(s) are only in src
	SrcGapRGVKs map[string]map[string][]schema.GroupVersionKind
	// DstGapRGVKs contains RGVKs where group is in both source and destination api-servers but version(s) are only in dst
	DstGapRGVKs map[string]map[string][]schema.GroupVersionKind
}

// Resource holds support information for a resource
type Resource struct {
	ResourceName  string
	Source        []schema.GroupVersionKind
	Destination   []schema.GroupVersionKind
	NamespaceList []string
	Support       string
}
