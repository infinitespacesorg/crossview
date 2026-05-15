const BASE = '/api/isp'

export async function fetchSites(env) {
  const r = await fetch(`${BASE}/environments/${env}/sites`)
  if (!r.ok) throw new Error(`fetchSites: ${r.status}`)
  const data = await r.json()
  return data.sites ?? []
}

export async function fetchNodes(site, zone) {
  const url = zone
    ? `${BASE}/sites/${site}/nodes?zone=${zone}`
    : `${BASE}/sites/${site}/nodes`
  const r = await fetch(url)
  if (!r.ok) throw new Error(`fetchNodes: ${r.status}`)
  const data = await r.json()
  return data.nodes ?? []
}

export async function fetchPen(site) {
  const r = await fetch(`${BASE}/sites/${site}/pen`)
  if (!r.ok) throw new Error(`fetchPen: ${r.status}`)
  const data = await r.json()
  return data.nodes ?? []
}

export async function musterNodes(nodeIds) {
  const r = await fetch(`${BASE}/nodes/muster`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ node_ids: nodeIds }),
  })
  if (!r.ok) throw new Error(`muster: ${r.status}`)
  return r.json()
}

export async function fetchFleetStatus() {
  const r = await fetch(`${BASE}/fleet/status`)
  if (!r.ok) throw new Error(`fleetStatus: ${r.status}`)
  return r.json()
}
