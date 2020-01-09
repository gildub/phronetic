package cluster_test

import (
	"testing"

	"github.com/gildub/phronetic/pkg/transform"
	"github.com/gildub/phronetic/pkg/transform/cluster"
	cpmatest "github.com/gildub/phronetic/pkg/transform/internal/test"
	o7tapiroute "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"

	k8sapicore "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReportPods(t *testing.T) {
	expectedPodRepors := make([]cluster.PodReport, 0)
	expectedPodRepors = append(expectedPodRepors, cluster.PodReport{Name: "test-pod1"})
	expectedPodRepors = append(expectedPodRepors, cluster.PodReport{Name: "test-pod2"})

	testCases := []struct {
		name              string
		inputPodList      *k8sapicore.PodList
		expectedPodRepors []cluster.PodReport
	}{
		{
			name:              "generate pod report",
			inputPodList:      cpmatest.CreateTestPodList(),
			expectedPodRepors: expectedPodRepors,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reportedNamespace := &cluster.NamespaceReport{}
			cluster.ReportPods(reportedNamespace, tc.inputPodList)
			assert.Equal(t, tc.expectedPodRepors, reportedNamespace.Pods)
		})
	}
}

func TestReportNamespaceResources(t *testing.T) {
	expectedCPU := &resource.Quantity{
		Format: resource.DecimalSI,
	}
	expectedCPU.Set(int64(2))
	expectedMemory := &resource.Quantity{
		Format: resource.BinarySI,
	}
	expectedMemory.Set(int64(2))

	expectedResources := &cluster.ContainerResourcesReport{
		ContainerCount: 2,
		CPUTotal:       expectedCPU,
		MemoryTotal:    expectedMemory,
	}

	testCases := []struct {
		name              string
		inputPodList      *k8sapicore.PodList
		expectedResources cluster.ContainerResourcesReport
	}{
		{
			name:              "generate resource report",
			inputPodList:      cpmatest.CreateTestPodResourceList(),
			expectedResources: *expectedResources,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reportedNamespace := &cluster.NamespaceReport{}
			cluster.ReportResources(reportedNamespace, tc.inputPodList)
			assert.Equal(t, tc.expectedResources, reportedNamespace.Resources)
		})
	}
}

func TestReportRoutes(t *testing.T) {
	alternateBackends := make([]o7tapiroute.RouteTargetReference, 0)
	alternateBackends = append(alternateBackends, o7tapiroute.RouteTargetReference{
		Kind: "testkind",
		Name: "testname",
	})

	to := o7tapiroute.RouteTargetReference{
		Kind: "testkindTo",
		Name: "testTo",
	}

	tls := &o7tapiroute.TLSConfig{
		Termination: o7tapiroute.TLSTerminationEdge,
	}

	expectedRouteReport := make([]cluster.RouteReport, 0)
	expectedRouteReport = append(expectedRouteReport, cluster.RouteReport{
		Name:              "route1",
		Host:              "testhost",
		Path:              "testpath",
		AlternateBackends: alternateBackends,
		TLS:               tls,
		To:                to,
		WildcardPolicy:    "None",
	})

	testCases := []struct {
		name                string
		inputRouteList      *o7tapiroute.RouteList
		expectedRouteReport []cluster.RouteReport
	}{
		{
			name:                "generate route report",
			inputRouteList:      cpmatest.CreateTestRouteList(),
			expectedRouteReport: expectedRouteReport,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reportedNamespace := &cluster.NamespaceReport{}
			cluster.ReportRoutes(reportedNamespace, tc.inputRouteList)
			assert.Equal(t, tc.expectedRouteReport, reportedNamespace.Routes)
		})
	}
}

func TestReportMisssingGVs(t *testing.T) {
	expectedMissingGVs := make([]cluster.NewGVsReport, 0)
	expectedMissingGVs = append(expectedMissingGVs, cluster.NewGVsReport{
		GroupVersion: "testgroupversion/v1",
	})

	testCases := []struct {
		name                  string
		inputGroupVersions    *metav1.APIGroupList
		inputDstGroupVersions *metav1.APIGroupList
		expectedNewGVs        []cluster.NewGVsReport
	}{
		{
			name:                  "generate missing groupversions report",
			inputGroupVersions:    cpmatest.CreateTestClusterGroupVersions("testgroupversion", "v1beta1"),
			inputDstGroupVersions: cpmatest.CreateTestClusterGroupVersions("testgroupversion", "v1"),
			expectedNewGVs:        expectedMissingGVs,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clusterNewGVs := &cluster.Report{}
			clusterNewGVs.ReportNewGVs(transform.NewGroupVersions(tc.inputGroupVersions, tc.inputDstGroupVersions))
			assert.Equal(t, tc.expectedNewGVs, clusterNewGVs.NewGVs)
		})
	}
}
