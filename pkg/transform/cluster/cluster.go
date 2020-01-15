package cluster

import (
	"sort"

	"github.com/gildub/phronetic/pkg/api"
	"github.com/sirupsen/logrus"

	"k8s.io/api/apps/v1beta1"
	k8sapicore "k8s.io/api/core/v1"
	k8scorev1 "k8s.io/api/core/v1"
	extv1b1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Report represents json report of k8s resources
type Report struct {
	Nodes            []NodeReport          `json:"nodes"`
	Namespace        NamespaceReport       `json:"namespace,omitempty"`
	PVs              []PVReport            `json:"pvs,omitempty"`
	StorageClasses   []StorageClassReport  `json:"storageClasses,omitempty"`
	OldGroupVersions []GroupVersionsReport `json:"oldGroupVersions,omitempty"`
	NewGroupVersions []GroupVersionsReport `json:"newGroupVersions,omitempty"`
}

// NodeReport represents json report of k8s nodes
type NodeReport struct {
	Name       string        `json:"name"`
	MasterNode bool          `json:"masterNode"`
	Resources  NodeResources `json:"resources"`
}

// NodeResources represents a json report of Node resources
type NodeResources struct {
	CPU            *resource.Quantity `json:"cpu"`
	MemoryConsumed *resource.Quantity `json:"memoryConsumed"`
	MemoryCapacity *resource.Quantity `json:"memoryCapacity"`
	RunningPods    *resource.Quantity `json:"runningPods"`
	PodCapacity    *resource.Quantity `json:"podCapacity"`
}

// NamespaceReport represents json report of k8s namespaces
type NamespaceReport struct {
	Name         string                   `json:"name"`
	LatestChange k8sMeta.Time             `json:"latestChange,omitempty"`
	Resources    ContainerResourcesReport `json:"resources,omitempty"`
	Pods         []PodReport              `json:"pods,omitempty"`
	DaemonSets   []DaemonSetReport        `json:"daemonSets,omitempty"`
	Deployments  []DeploymentReport       `json:"deployments,omitempty"`
	Quotas       []ResourceQuotaReport    `json:"quotas,omitempty"`
	PVCs         []PVСReport              `json:"persistentVolumeClaims,omitempty"`
}

// PodReport represents json report of k8s pods
type PodReport struct {
	Name string `json:"name"`
}

// ResourceQuotaReport represents json report of Quota resources
type ResourceQuotaReport struct {
	Name          string                          `json:"name"`
	Hard          k8scorev1.ResourceList          `json:"hard,omitempty"`
	ScopeSelector *k8sapicore.ScopeSelector       `json:"selector,omitempty"`
	Scopes        []k8sapicore.ResourceQuotaScope `json:"scopes,omitempty"`
}

// DaemonSetReport represents json report of k8s DaemonSet relevant information
type DaemonSetReport struct {
	Name         string       `json:"name"`
	LatestChange k8sMeta.Time `json:"latestChange,omitempty"`
}

// DeploymentReport represents json report of DeploymentReport resources
type DeploymentReport struct {
	Name         string       `json:"name"`
	LatestChange k8sMeta.Time `json:"latestChange,omitempty"`
}

// ContainerResourcesReport represents json report for aggregated container resources
type ContainerResourcesReport struct {
	ContainerCount int                `json:"containerCount"`
	CPUTotal       *resource.Quantity `json:"cpuTotal"`
	MemoryTotal    *resource.Quantity `json:"memoryTotal"`
}

// PVReport represents json report of k8s PVs
type PVReport struct {
	Name          string                                   `json:"name"`
	Driver        k8sapicore.PersistentVolumeSource        `json:"driver"`
	StorageClass  string                                   `json:"storageClass,omitempty"`
	Capacity      k8sapicore.ResourceList                  `json:"capacity,omitempty"`
	Phase         k8sapicore.PersistentVolumePhase         `json:"phase,omitempty"`
	ReclaimPolicy k8sapicore.PersistentVolumeReclaimPolicy `json:"persistentVolumeReclaimPolicy,omitempty" protobuf:"bytes,5,opt,name=persistentVolumeReclaimPolicy,casttype=PersistentVolumeReclaimPolicy"`
}

// StorageClassReport represents json report of k8s storage classes
type StorageClassReport struct {
	Name        string `json:"name"`
	Provisioner string `json:"provisioner"`
}

// GroupVersionsReport represents json report of k8s storage classes
type GroupVersionsReport struct {
	GroupVersion string `json:"groupversion"`
}

// PVСReport represents json report of k8s PVs
type PVСReport struct {
	Name          string                                   `json:"name"`
	PVName        string                                   `json:"pvname"`
	AccessModes   []k8scorev1.PersistentVolumeAccessMode   `json:"accessModes,omitempty" protobuf:"bytes,1,rep,name=accessModes,casttype=PersistentVolumeAccessMode"`
	StorageClass  string                                   `json:"storageClass"`
	Capacity      k8sapicore.ResourceList                  `json:"capacity,omitempty"`
	ReclaimPolicy k8sapicore.PersistentVolumeReclaimPolicy `json:"persistentVolumeReclaimPolicy,omitempty" protobuf:"bytes,5,opt,name=persistentVolumeReclaimPolicy,casttype=PersistentVolumeReclaimPolicy"`
}

// GenClusterReport inserts report values into structures for json output
func GenClusterReport(apiResources api.Resources) (clusterReport Report) {
	clusterReport.ReportNamespaceResources(apiResources.NamespaceResources)
	clusterReport.ReportOldGVs(apiResources.OldGroupVersions)
	clusterReport.ReportNewGVs(apiResources.NewGroupVersions)
	return
}

// ReportContainerResources create report about container resources
func ReportContainerResources(reportedNamespace *NamespaceReport, pod *k8sapicore.Pod) {
	cpuTotal := reportedNamespace.Resources.CPUTotal.Value()
	memoryTotal := reportedNamespace.Resources.MemoryTotal.Value()

	for _, container := range pod.Spec.Containers {
		cpuTotal += container.Resources.Requests.Cpu().Value()
		memoryTotal += container.Resources.Requests.Memory().Value()
	}
	reportedNamespace.Resources.CPUTotal.Set(cpuTotal)
	reportedNamespace.Resources.MemoryTotal.Set(memoryTotal)
	reportedNamespace.Resources.ContainerCount += len(pod.Spec.Containers)
}

// ReportDaemonSets generate DaemonSet report
func ReportDaemonSets(reporeportedNamespace *NamespaceReport, dsList *extv1b1.DaemonSetList) {
	for _, ds := range dsList.Items {
		reportedDS := DaemonSetReport{
			Name:         ds.Name,
			LatestChange: ds.ObjectMeta.CreationTimestamp,
		}

		reporeportedNamespace.DaemonSets = append(reporeportedNamespace.DaemonSets, reportedDS)
	}
}

// ReportDeployments generate Deployments report
func ReportDeployments(reportedNamespace *NamespaceReport, deploymentList *v1beta1.DeploymentList) {
	for _, deployment := range deploymentList.Items {
		reportedDeployment := DeploymentReport{
			Name:         deployment.Name,
			LatestChange: deployment.ObjectMeta.CreationTimestamp,
		}

		reportedNamespace.Deployments = append(reportedNamespace.Deployments, reportedDeployment)
	}
}

// ReportNamespaceResources fills in information about resources of a namespace
func (clusterReport *Report) ReportNamespaceResources(apiResources *api.NamespaceResources) {
	logrus.Info("ClusterReport::ReportNamespaceResources")

	reportedNamespace := NamespaceReport{Name: apiResources.Namespace.Name}

	ReportResourceQuotas(&reportedNamespace, apiResources.ResourceQuotaList)
	ReportPods(&reportedNamespace, apiResources.PodList)
	ReportResources(&reportedNamespace, apiResources.PodList)
	ReportDeployments(&reportedNamespace, apiResources.DeploymentList)
	ReportDaemonSets(&reportedNamespace, apiResources.DaemonSetList)
	ReportPVCs(&reportedNamespace, apiResources.PVCList, clusterReport.PVs)
	clusterReport.Namespace = reportedNamespace
}

// ReportNodeResources parse and insert info about consumed resources
func ReportNodeResources(repotedNode *NodeReport, nodeStatus k8sapicore.NodeStatus, apiResources api.Resources) {
	repotedNode.Resources.CPU = nodeStatus.Capacity.Cpu()

	repotedNode.Resources.MemoryCapacity = nodeStatus.Capacity.Memory()

	memConsumed := new(resource.Quantity)
	memCapacity, _ := nodeStatus.Capacity.Memory().AsInt64()
	memAllocatable, _ := nodeStatus.Allocatable.Memory().AsInt64()
	memConsumed.Set(memCapacity - memAllocatable)
	memConsumed.Format = resource.BinarySI
	repotedNode.Resources.MemoryConsumed = memConsumed

	var runningPodsCount int64
	for _, pod := range apiResources.NamespaceResources.PodList.Items {
		if pod.Spec.NodeName == repotedNode.Name {
			runningPodsCount++
		}
	}

	podsRunning := new(resource.Quantity)
	podsRunning.Set(runningPodsCount)
	podsRunning.Format = resource.DecimalSI
	repotedNode.Resources.RunningPods = podsRunning

	repotedNode.Resources.PodCapacity = nodeStatus.Capacity.Pods()
}

// ReportPods creates info about cluster pods
func ReportPods(reportedNamespace *NamespaceReport, podList *k8sapicore.PodList) {
	for _, pod := range podList.Items {
		reportedPod := PodReport{Name: pod.Name}
		reportedNamespace.Pods = append(reportedNamespace.Pods, reportedPod)

		// Update namespace touch timestamp
		if pod.ObjectMeta.CreationTimestamp.Time.Unix() > reportedNamespace.LatestChange.Time.Unix() {
			reportedNamespace.LatestChange = pod.ObjectMeta.CreationTimestamp
		}
	}
}

// ReportResourceQuotas creates report about quotas
func ReportResourceQuotas(reportedNamespace *NamespaceReport, quotaList *k8sapicore.ResourceQuotaList) {
	for _, quota := range quotaList.Items {
		reportedQuota := ResourceQuotaReport{
			Name:          quota.ObjectMeta.Name,
			Hard:          quota.Spec.Hard,
			ScopeSelector: quota.Spec.ScopeSelector,
			Scopes:        quota.Spec.Scopes,
		}
		reportedNamespace.Quotas = append(reportedNamespace.Quotas, reportedQuota)
	}
}

// ReportResources create report about namespace resources
func ReportResources(reportedNamespace *NamespaceReport, podList *k8sapicore.PodList) {
	resources := ContainerResourcesReport{
		CPUTotal:    &resource.Quantity{Format: resource.DecimalSI},
		MemoryTotal: &resource.Quantity{Format: resource.BinarySI},
	}
	reportedNamespace.Resources = resources

	for _, pod := range podList.Items {
		ReportContainerResources(reportedNamespace, &pod)
	}
}

// ReportPVCs generate PVC report
func ReportPVCs(reporeportedNamespace *NamespaceReport, pvcList *k8scorev1.PersistentVolumeClaimList, pvList []PVReport) {
	for _, pvc := range pvcList.Items {
		if len(pvList) == 0 {
			return
		}
		idx := sort.Search(len(pvList), func(i int) bool {
			return pvList[i].Name >= pvc.Spec.VolumeName
		})
		pv := pvList[idx]

		var storageClass string
		if pvc.Spec.StorageClassName != nil {
			storageClass = *pvc.Spec.StorageClassName
		} else {
			storageClass = "None"
		}

		reportedPVC := PVСReport{
			Name:          pvc.Name,
			PVName:        pvc.Spec.VolumeName,
			AccessModes:   pvc.Spec.AccessModes,
			StorageClass:  storageClass,
			Capacity:      pv.Capacity,
			ReclaimPolicy: pv.ReclaimPolicy,
		}

		reporeportedNamespace.PVCs = append(reporeportedNamespace.PVCs, reportedPVC)
	}
}

// ReportOldGVs reports GroupVersion present in source but in destination cluster
func (clusterReport *Report) ReportOldGVs(list []string) {
	logrus.Info("ClusterReport::ReportOldGroupVersions")
	for _, groupVersion := range list {
		reportedGroupVersions := GroupVersionsReport{
			GroupVersion: groupVersion,
		}

		clusterReport.OldGroupVersions = append(clusterReport.OldGroupVersions, reportedGroupVersions)
	}
}

// ReportNewGVs reports GroupVersion present in destination but in source cluster
func (clusterReport *Report) ReportNewGVs(list []string) {
	logrus.Info("ClusterReport::ReportNewGroupVersions")
	for _, groupVersion := range list {
		reportedGroupVersions := GroupVersionsReport{
			GroupVersion: groupVersion,
		}

		clusterReport.NewGroupVersions = append(clusterReport.NewGroupVersions, reportedGroupVersions)
	}
}

func deduplicate(s []string) []string {
	if len(s) <= 1 {
		return s
	}

	result := []string{}
	seen := make(map[string]struct{})
	for _, val := range s {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}
