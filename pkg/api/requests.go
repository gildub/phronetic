package api

import (
	"context"

	migv1alpha1 "github.com/fusor/mig-controller/pkg/apis/migration/v1alpha1"
	"github.com/sirupsen/logrus"

	"k8s.io/api/apps/v1beta1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Resources represent api resources used in report
type Resources struct {
	HPAList              *autoscalingv1.HorizontalPodAutoscalerList
	SrcGroupVersions     *metav1.APIGroupList
	DstGroupVersions     *metav1.APIGroupList
	NodeList             *corev1.NodeList
	PersistentVolumeList *corev1.PersistentVolumeList
	StorageClassList     *storagev1.StorageClassList
	OldGroupVersions     []string
	NewGroupVersions     []string
	NamespaceResources   *NamespaceResources
}

// NamespaceResources holds all resources that belong to a namespace
type NamespaceResources struct {
	Namespace         *corev1.Namespace
	DaemonSetList     *extv1beta1.DaemonSetList
	DeploymentList    *v1beta1.DeploymentList
	PodList           *corev1.PodList
	ResourceQuotaList *corev1.ResourceQuotaList
	PVCList           *corev1.PersistentVolumeClaimList
}

var listOptions metav1.ListOptions
var getOptions metav1.GetOptions

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

// ListDaemonSets will collect all DS from specific namespace
func ListDaemonSets(client *kubernetes.Clientset, namespace string) *extv1beta1.DaemonSetList {
	daemonSets, err := client.ExtensionsV1beta1().DaemonSets(namespace).List(listOptions)
	if err != nil {
		logrus.Fatal(err)
	}
	return daemonSets
}

// ListDeployments will list all deployments seeding in the selected namespace
func ListDeployments(client *kubernetes.Clientset, namespace string) *v1beta1.DeploymentList {
	deployments, err := client.AppsV1beta1().Deployments(namespace).List(listOptions)
	if err != nil {
		logrus.Fatal(err)
	}
	return deployments
}

// ListGroupVersions list all GV
func ListGroupVersions(client *kubernetes.Clientset) *metav1.APIGroupList {
	groupVersions, err := client.ServerGroups()
	if err != nil {
		logrus.Fatal(err)
	}
	return groupVersions
}

// ListHPAs gets Horizontal Pod Autoscaler objects
func ListHPAs(client *kubernetes.Clientset, namespace string, ch chan<- *autoscalingv1.HorizontalPodAutoscalerList) {
	hpas, err := client.AutoscalingV1().HorizontalPodAutoscalers(namespace).List(listOptions)
	if err != nil {
		logrus.Fatal(err)
	}
	ch <- hpas
}

// ListNamespaces list all namespaces, wrapper around client-go
func ListNamespaces(client *kubernetes.Clientset, ch chan<- *corev1.NamespaceList) {
	namespaces, err := client.CoreV1().Namespaces().List(listOptions)
	if err != nil {
		logrus.Fatal(err)
	}
	ch <- namespaces
}

// ListNodes list all nodes, wrapper around client-go
func ListNodes(client *kubernetes.Clientset, ch chan<- *corev1.NodeList) {
	nodes, err := client.CoreV1().Nodes().List(listOptions)
	if err != nil {
		logrus.Fatal(err)
	}
	ch <- nodes
}

// ListPods list all pods in namespace, wrapper around client-go
func ListPods(client *kubernetes.Clientset, namespace string) *corev1.PodList {
	pods, err := client.CoreV1().Pods(namespace).List(listOptions)
	if err != nil {
		logrus.Fatal(err)
	}
	return pods
}

// ListResourceQuotas list all quotas classes, wrapper around client-go
func ListResourceQuotas(client *kubernetes.Clientset, namespace string) *corev1.ResourceQuotaList {
	quotas, err := client.CoreV1().ResourceQuotas(namespace).List(listOptions)
	if err != nil {
		logrus.Fatal(err)
	}
	return quotas
}

// ListStorageClasses list all storage classes, wrapper around client-go
func ListStorageClasses(client *kubernetes.Clientset, ch chan<- *storagev1.StorageClassList) {
	sc, err := client.StorageV1().StorageClasses().List(listOptions)
	if err != nil {
		logrus.Fatal(err)
	}
	ch <- sc
}

// ListPVs list all PVs, wrapper around client-go
func ListPVs(client *kubernetes.Clientset, ch chan<- *corev1.PersistentVolumeList) {
	pvs, err := client.CoreV1().PersistentVolumes().List(listOptions)
	if err != nil {
		logrus.Fatal(err)
	}
	ch <- pvs
}

// ListPVCs list all PVs, wrapper around client-go
func ListPVCs(client *kubernetes.Clientset, namespace string) *corev1.PersistentVolumeClaimList {
	pvcs, err := client.CoreV1().PersistentVolumeClaims(namespace).List(listOptions)
	if err != nil {
		logrus.Fatal(err)
	}
	return pvcs
}
