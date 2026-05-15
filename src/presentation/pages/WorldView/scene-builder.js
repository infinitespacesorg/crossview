import { ISP_THEME } from 'isp-fossflow'

const ROLE_ICON = {
  'capture-node':   '/icons/capture-node.svg',
  'display-node':   '/icons/display-node.svg',
  'compute-node':   '/icons/compute-node.svg',
  'serve-node':     '/icons/serve-node.svg',
  'relay-node':     '/icons/relay-node.svg',
  'sensor-node':    '/icons/sensor-node.svg',
  'gateway-node':   '/icons/gateway-node.svg',
  'interface-node': '/icons/interface-node.svg',
}

const ZONE_COLOR = {
  edge:    'zone-edge',
  compute: 'zone-compute',
  gpu:     'zone-gpu',
  static:  'zone-static',
}

function buildIcons(nodes) {
  const seen = new Set()
  return nodes
    .filter(n => { if (seen.has(n.role)) return false; seen.add(n.role); return true })
    .map(n => ({ id: n.role, name: n.role, url: ROLE_ICON[n.role] ?? '/icons/compute-node.svg', isIsometric: false }))
}

function gridTiles(nodes, originX, originY, cols) {
  return nodes.map((node, i) => ({
    id: node.id,
    tile: { x: originX + 1 + (i % cols), y: originY + 1 + Math.floor(i / cols) },
  }))
}

export function buildSiteScene(siteName, nodes, env) {
  const byZone = {}
  for (const n of nodes) {
    if (!byZone[n.zone]) byZone[n.zone] = []
    byZone[n.zone].push(n)
  }

  const zones = Object.keys(byZone)
  const ZONE_WIDTH = 6, ZONE_GAP = 2, ZONE_HEIGHT = 7
  const rectangles = [], viewItems = [], textBoxes = [], items = []

  zones.forEach((zone, zi) => {
    const originX = zi * (ZONE_WIDTH + ZONE_GAP)
    rectangles.push({
      id: `r-${zone}`,
      color: ZONE_COLOR[zone] ?? 'earth',
      from: { x: originX, y: 0 },
      to: { x: originX + ZONE_WIDTH, y: ZONE_HEIGHT },
    })
    textBoxes.push({ id: `lbl-${zone}`, tile: { x: originX, y: 0 }, content: zone.toUpperCase() + ' ZONE', fontSize: 0.4 })
    viewItems.push(...gridTiles(byZone[zone], originX, 0, 3))
  })

  textBoxes.push({ id: 'lbl-site', tile: { x: 0, y: ZONE_HEIGHT + 1 }, content: `${siteName.toUpperCase()} · ${env}`, fontSize: 0.5 })

  for (const node of nodes) {
    items.push({
      id: node.id,
      name: node.name,
      description: `**${node.role}**\n\nSite: ${node.site}\nZone: ${node.zone}\nHerd: ${node.herd || '—'}\nPhase: ${node.phase}`,
      icon: node.role,
    })
  }

  return {
    icons: buildIcons(nodes),
    colors: ISP_THEME.colors,
    items,
    views: [{
      id: `${siteName}-site`,
      name: `${siteName} — Site View`,
      description: `Environment: ${env} · Site: ${siteName} · ${zones.length} zones · ${nodes.length} nodes`,
      rectangles,
      items: viewItems,
      connectors: [],
      textBoxes,
    }],
  }
}

export function buildRangeScene(sites) {
  const items = sites.map(s => ({
    id: s.id,
    name: s.name,
    description: `**${s.name}**\n\n${s.node_count} nodes · ${s.zones?.length ?? 0} zones\nStatus: ${s.fleet_status}`,
    icon: 'site-outpost',
  }))

  const viewItems = sites.map((s, i) => ({ id: s.id, tile: { x: i * 4, y: 0 } }))

  return {
    icons: [{ id: 'site-outpost', name: 'Site', url: '/icons/site-outpost.svg', isIsometric: false }],
    colors: ISP_THEME.colors,
    items,
    views: [{
      id: 'range',
      name: 'Range View — All Sites',
      description: `${sites.length} sites`,
      rectangles: [],
      items: viewItems,
      connectors: [],
      textBoxes: [],
    }],
  }
}
