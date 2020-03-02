package test

import (
	"fmt"

	k8sapicore "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateTestPVList create test pv list
func CreateTestPVList() *k8sapicore.PersistentVolumeList {
	pvList := &k8sapicore.PersistentVolumeList{}
	pvList.Items = make([]k8sapicore.PersistentVolume, 1)

	resources := make(k8sapicore.ResourceList)
	cpu := resource.Quantity{
		Format: resource.DecimalSI,
	}
	cpu.Set(int64(1))
	resources["cpu"] = cpu

	memory := resource.Quantity{
		Format: resource.BinarySI,
	}
	memory.Set(int64(1))
	resources["memory"] = memory

	driver := k8sapicore.PersistentVolumeSource{
		NFS: &k8sapicore.NFSVolumeSource{
			Server: "example.com",
		},
	}

	pvList.Items[0] = k8sapicore.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testpv",
		},
		Spec: k8sapicore.PersistentVolumeSpec{
			PersistentVolumeSource:        driver,
			StorageClassName:              "testclass",
			Capacity:                      resources,
			PersistentVolumeReclaimPolicy: k8sapicore.PersistentVolumeReclaimPolicy("testpolicy"),
		},
		Status: k8sapicore.PersistentVolumeStatus{
			Phase: k8sapicore.VolumePending,
		},
	}

	return pvList
}

// CreateTestClusterGroupVersions test for GroupVersionList
func CreateTestClusterGroupVersions(group string, version string) *metav1.APIGroupList {
	groupList := &metav1.APIGroupList{}
	groupList.Groups = make([]metav1.APIGroup, 1)

	groupList.Groups[0] = metav1.APIGroup{
		TypeMeta: v1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		Name: group,
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: fmt.Sprintf("%s/%s", group, version),
				Version: version},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: fmt.Sprintf("%s/%s", group, version),
			Version:      version,
		},
	}

	return groupList
}
