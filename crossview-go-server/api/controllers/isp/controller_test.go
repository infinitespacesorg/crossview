package isp_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"crossview-go-server/api/controllers/isp"
)

func makeUnstructuredNode(name, site, tier, phase, role string) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "infinitespaces.org/v1alpha1",
			"kind":       "ManagedNode",
			"metadata": map[string]any{
				"name": name,
				"labels": map[string]any{
					"isp/site": site,
					"isp/tier": tier,
				},
				"annotations": map[string]any{
					"isp/node-role": role,
				},
				"uid": name + "-uid",
			},
			"status": map[string]any{
				"phase": phase,
			},
		},
	}
}

func TestAggregateSites(t *testing.T) {
	nodes := []unstructured.Unstructured{
		makeUnstructuredNode("a", "studio", "edge", "Ready", "compute-node"),
		makeUnstructuredNode("b", "studio", "compute", "Ready", "relay-node"),
		makeUnstructuredNode("c", "venue-paris", "edge", "Provisioning", "capture-node"),
	}
	sites := isp.AggregateSites(nodes)
	if len(sites) != 2 {
		t.Fatalf("expected 2 sites, got %d", len(sites))
	}
	var studio *isp.Site
	for i := range sites {
		if sites[i].Name == "studio" {
			studio = &sites[i]
		}
	}
	if studio == nil {
		t.Fatal("studio site not found")
	}
	if studio.NodeCount != 2 {
		t.Fatalf("expected 2 nodes in studio, got %d", studio.NodeCount)
	}
}

func TestNodesToISP(t *testing.T) {
	nodes := []unstructured.Unstructured{
		makeUnstructuredNode("node-1", "studio", "edge", "Provisioning", "capture-node"),
	}
	ispNodes := isp.NodesToISP(nodes)
	if len(ispNodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(ispNodes))
	}
	n := ispNodes[0]
	if n.Phase != "Provisioning" {
		t.Errorf("expected phase Provisioning, got %s", n.Phase)
	}
	if n.Role != "capture-node" {
		t.Errorf("expected role capture-node, got %s", n.Role)
	}
}

// Ensure metav1 import is used (keeps the build clean in test files that need it later)
var _ = metav1.ListOptions{}
