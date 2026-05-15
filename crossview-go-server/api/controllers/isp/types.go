package isp

// Site is an IS site — a named physical or logical location.
// Derived from the isp/site label on ManagedNode CRs.
type Site struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Zones     []Zone `json:"zones"`
	NodeCount int    `json:"node_count"`
	Status    string `json:"fleet_status"` // "ok" | "degraded" | "offline"
}

// Zone is a compute-typed area within a Site.
// Derived from the isp/tier label: edge | compute | gpu | static.
type Zone struct {
	Name      string `json:"name"`
	NodeCount int    `json:"node_count"`
}

// Node is an individual PlayBox instance.
// Phase reflects the deployed CRD enum:
// Provisioning | Ready | Assigned | Active | Offline | Upgrading | Maintenance
// IS mapping: Provisioning → Pen, Ready → Pasture, Assigned → Herd.
type Node struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Role   string `json:"role"`   // capture-node, display-node, etc. from isp/node-role annotation
	Site   string `json:"site"`
	Zone   string `json:"zone"`   // edge | compute | gpu | static
	Herd   string `json:"herd"`   // NodeSet name, empty if unassigned
	Phase  string `json:"phase"`  // Provisioning | Ready | Assigned | Active | Offline | ...
	Status string `json:"status"` // "online" | "offline"
}

// Herd is a named group of nodes sharing a composition (NodeSet CR).
type Herd struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Nodes       []string `json:"nodes"`
	Composition string   `json:"composition"`
	Status      string   `json:"status"`
}

// FleetStatus is an aggregate health snapshot.
type FleetStatus struct {
	Total  int            `json:"total"`
	Online int            `json:"online"`
	Staged int            `json:"staged"`  // phase == Provisioning
	InHerd int            `json:"in_herd"` // phase == Assigned
	BySite map[string]int `json:"by_site"`
	ByRole map[string]int `json:"by_role"`
}

// MusterRequest approves staged nodes (stub — actual operation is Headscale ACL; pending CRD reconciliation).
type MusterRequest struct {
	NodeIDs []string `json:"node_ids"`
}

// MusterResponse reports which nodes were approved.
type MusterResponse struct {
	Approved []string `json:"approved"`
	Failed   []string `json:"failed"`
}

// CreateHerdRequest creates or updates a NodeSet CR.
type CreateHerdRequest struct {
	Name        string   `json:"name"`
	NodeIDs     []string `json:"node_ids"`
	Composition string   `json:"composition"`
}
