package api

import (
	"context"

	migv1alpha1 "github.com/fusor/mig-controller/pkg/apis/migration/v1alpha1"
	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var listOptions metav1.ListOptions
var getOptions metav1.GetOptions

// RESTMapperGetGRs lists all GVKs for a resource
func RESTMapperGetGRs(client *kubernetes.Clientset) meta.RESTMapper {
	groupResources, err := restmapper.GetAPIGroupResources(client.Discovery())
	if err != nil {
		logrus.Fatal(err)
	}
	return restmapper.NewDiscoveryRESTMapper(groupResources)
}

// GetKindsFor lists all GVKs for a resource
func GetKindsFor(restMapper meta.RESTMapper, resource string) []schema.GroupVersionKind {
	gvr := schema.GroupVersionResource{Group: "", Version: "", Resource: resource}
	gvks, err := restMapper.KindsFor(gvr)
	if err != nil {
		logrus.Fatal(err)
	}
	return gvks
}

// ListServerResources list all resources
func ListServerResources(client *kubernetes.Clientset) []*metav1.APIResourceList {
	resources, err := client.ServerResources()
	if err != nil {
		logrus.Fatal(err)
	}
	return resources
}

// GetMigCluster get MigrationCluster
func GetMigCluster(client ctrlclient.Client, name string) migv1alpha1.MigCluster {
	objectKey := types.NamespacedName{
		Namespace: "openshift-migration",
		Name:      name,
	}

	migCluster := migv1alpha1.MigCluster{}
	err := client.Get(context.TODO(), objectKey, &migCluster)
	if err != nil {
		logrus.Fatal(err)
	}
	return migCluster
}

// GetMigPlan get MigrationPlan
func GetMigPlan(client ctrlclient.Client, name string) (migv1alpha1.MigPlan, error) {
	objectKey := types.NamespacedName{
		Namespace: "openshift-migration",
		Name:      name,
	}

	migPlan := migv1alpha1.MigPlan{}
	err := client.Get(context.TODO(), objectKey, &migPlan)
	return migPlan, err
}

// GetNamespace get namespace
func GetNamespace(client *kubernetes.Clientset, name string) *corev1.Namespace {
	namespace, err := client.CoreV1().Namespaces().Get(name, getOptions)
	if err != nil {
		logrus.Fatal(err)
	}
	return namespace
}
