package isp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"crossview-go-server/lib"
	"crossview-go-server/services"
)

// ISP CRD GVRs — group confirmed from product/components/control-plane/manifests/managednode/managednode-crd.yaml
var (
	managedNodeGVR = schema.GroupVersionResource{
		Group: "infinitespaces.org", Version: "v1alpha1", Resource: "managednodes",
	}
	nodeSetGVR = schema.GroupVersionResource{
		Group: "infinitespaces.org", Version: "v1alpha1", Resource: "nodesets",
	}
)

// ISP label and annotation keys
const (
	labelSite = "isp/site"
	labelTier = "isp/tier"
	annotRole = "isp/node-role"
)

type ISPController struct {
	logger            lib.Logger
	kubernetesService services.KubernetesServiceInterface
}

func NewISPController(logger lib.Logger, kubernetesService services.KubernetesServiceInterface) ISPController {
	return ISPController{logger: logger, kubernetesService: kubernetesService}
}

// dynClient builds a dynamic client from the current K8s config.
// Called per-handler so it always uses the current active context.
func (c *ISPController) dynClient() (dynamic.Interface, error) {
	cfg, err := c.kubernetesService.GetConfig()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(cfg)
}

// AggregateSites groups a list of ManagedNode unstructured objects by isp/site label.
// Exported for testing.
func AggregateSites(nodes []unstructured.Unstructured) []Site {
	siteZones := map[string]map[string]int{}
	for _, n := range nodes {
		labels := n.GetLabels()
		site := labels[labelSite]
		if site == "" {
			site = "unknown"
		}
		zone := labels[labelTier]
		if _, ok := siteZones[site]; !ok {
			siteZones[site] = map[string]int{}
		}
		siteZones[site][zone]++
	}
	sites := make([]Site, 0, len(siteZones))
	for name, zones := range siteZones {
		total := 0
		zoneList := make([]Zone, 0, len(zones))
		for z, c := range zones {
			zoneList = append(zoneList, Zone{Name: z, NodeCount: c})
			total += c
		}
		sites = append(sites, Site{
			ID: name, Name: name,
			Zones: zoneList, NodeCount: total, Status: "ok",
		})
	}
	return sites
}

// NodesToISP converts unstructured ManagedNode objects to IS Node structs.
// Exported for testing.
func NodesToISP(nodes []unstructured.Unstructured) []Node {
	result := make([]Node, 0, len(nodes))
	for _, item := range nodes {
		labels := item.GetLabels()
		ann := item.GetAnnotations()
		phase, _, _ := unstructured.NestedString(item.Object, "status", "phase")
		status := "online"
		if phase == "Offline" {
			status = "offline"
		}
		result = append(result, Node{
			ID:     string(item.GetUID()),
			Name:   item.GetName(),
			Role:   ann[annotRole],
			Site:   labels[labelSite],
			Zone:   labels[labelTier],
			Phase:  phase,
			Status: status,
		})
	}
	return result
}

// ListSites handles GET /api/isp/environments/:env/sites
// ManagedNode is cluster-scoped — list with Namespace("") = all.
func (c *ISPController) ListSites(ctx *gin.Context) {
	dyn, err := c.dynClient()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	list, err := dyn.Resource(managedNodeGVR).Namespace("").List(ctx.Request.Context(), metav1.ListOptions{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sites := AggregateSites(list.Items)
	ctx.JSON(http.StatusOK, gin.H{"sites": sites, "count": len(sites)})
}

// ListNodes handles GET /api/isp/sites/:site/nodes?zone=<tier>
func (c *ISPController) ListNodes(ctx *gin.Context) {
	dyn, err := c.dynClient()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	site := ctx.Param("site")
	zone := ctx.Query("zone")

	selector := labelSite + "=" + site
	if zone != "" {
		selector += "," + labelTier + "=" + zone
	}

	list, err := dyn.Resource(managedNodeGVR).Namespace("").List(ctx.Request.Context(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nodes := NodesToISP(list.Items)
	ctx.JSON(http.StatusOK, gin.H{"nodes": nodes, "count": len(nodes)})
}

// ListPen handles GET /api/isp/sites/:site/pen — Provisioning-phase nodes (IS: Pen)
func (c *ISPController) ListPen(ctx *gin.Context) {
	dyn, err := c.dynClient()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	site := ctx.Param("site")
	list, err := dyn.Resource(managedNodeGVR).Namespace("").List(ctx.Request.Context(), metav1.ListOptions{
		LabelSelector: labelSite + "=" + site,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var penNodes []unstructured.Unstructured
	for _, item := range list.Items {
		phase, _, _ := unstructured.NestedString(item.Object, "status", "phase")
		if phase == "Provisioning" {
			penNodes = append(penNodes, item)
		}
	}

	nodes := NodesToISP(penNodes)
	ctx.JSON(http.StatusOK, gin.H{"nodes": nodes, "count": len(nodes)})
}

// Muster handles POST /api/isp/nodes/muster
// NOTE: The actual muster operation (Headscale ACL approval) requires strata-fabric integration
// pending control-plane CRD reconciliation (see retro-2026-04-30-fleet-crd-verification-skip.md).
// This endpoint returns 501 until that reconciliation is complete.
func (c *ISPController) Muster(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{
		"error": "muster not yet implemented — pending CRD reconciliation (CP-M6)",
	})
}

// CreateHerd handles POST /api/isp/herds
// NOTE: Returns 501 pending CRD reconciliation.
func (c *ISPController) CreateHerd(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{
		"error": "create herd not yet implemented — pending CRD reconciliation (CP-M6)",
	})
}

// RecallHerd handles DELETE /api/isp/herds/:herd
// NOTE: Returns 501 pending CRD reconciliation.
func (c *ISPController) RecallHerd(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, gin.H{
		"error": "recall not yet implemented — pending CRD reconciliation (CP-M6)",
	})
}

// FleetStatus handles GET /api/isp/fleet/status
func (c *ISPController) FleetStatus(ctx *gin.Context) {
	dyn, err := c.dynClient()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	list, err := dyn.Resource(managedNodeGVR).Namespace("").List(ctx.Request.Context(), metav1.ListOptions{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fs := FleetStatus{
		BySite: map[string]int{},
		ByRole: map[string]int{},
	}
	for _, item := range list.Items {
		fs.Total++
		labels := item.GetLabels()
		ann := item.GetAnnotations()
		phase, _, _ := unstructured.NestedString(item.Object, "status", "phase")

		if phase == "Provisioning" {
			fs.Staged++
		} else {
			fs.Online++
		}
		if phase == "Assigned" {
			fs.InHerd++
		}
		fs.BySite[labels[labelSite]]++
		fs.ByRole[ann[annotRole]]++
	}

	ctx.JSON(http.StatusOK, fs)
}

// ensure nodeSetGVR is referenced to avoid unused variable compile error
var _ = nodeSetGVR
