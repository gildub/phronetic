package api

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

/*
RGVKs is short for Resource -> Group - GroupVersionKind
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
	NamespaceResources *NamespaceResources
	// SrcRGVKs contains all RGVKs available on source api-server (trimmed of "/.*"" suffixes)
	SrcRGVKs map[string]map[string][]schema.GroupVersionKind
	// DstRGVKs contains all RGVKs available on destination api-server (trimmed of "/.* suffixes)
	DstRGVKs map[string]map[string][]schema.GroupVersionKind
	// SrcOnlyRGVKs contains RGVKs only available on source api-server
	SrcOnlyRGVKs map[string]map[string][]schema.GroupVersionKind
	// SrcGapRGVKs contains RGVKs which group is in both source and destination api-servers but version(s) are only in src
	SrcGapRGVKs map[string]map[string][]schema.GroupVersionKind
	// DstGapRGVKs contains RGVKs which group is in both source and destination api-servers but version(s) are only in dst
	DstGapRGVKs map[string]map[string][]schema.GroupVersionKind
}

// NamespaceResources holds all resources that belong to a namespace
type NamespaceResources struct {
	Namespace *corev1.Namespace
}
