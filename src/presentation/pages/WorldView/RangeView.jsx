import { useEffect, useState } from 'react'
import { Box, HStack, Text, Button, Badge } from '@chakra-ui/react'
import { ISPIsoflow } from 'isp-fossflow'
import { fetchSites, fetchFleetStatus } from './api.js'
import { buildRangeScene } from './scene-builder.js'

const STATUS_COLOR = { ok: 'green', degraded: 'orange', offline: 'red' }

export function RangeView({ env, onSiteSelect }) {
  const [scene, setScene] = useState(null)
  const [sites, setSites] = useState([])
  const [status, setStatus] = useState(null)
  const [error, setError] = useState(null)

  useEffect(() => {
    Promise.all([fetchSites(env), fetchFleetStatus()])
      .then(([s, fs]) => { setSites(s); setScene(buildRangeScene(s)); setStatus(fs) })
      .catch(e => setError(e.message))
  }, [env])

  if (error) return <Box p={4} color="red.400">Error: {error}</Box>
  if (!scene) return <Box p={4} color="blue.300">Loading fleet…</Box>

  return (
    <Box display="flex" flexDirection="column" width="100%" height="100%">
      {/* Fleet status bar */}
      {status && (
        <HStack px={4} py={1} bg="gray.950" borderBottom="1px solid" borderColor="gray.800" spacing={5} fontSize="xs">
          <Text color="gray.400">{status.total} nodes</Text>
          <Text color="blue.300">{status.online} online</Text>
          {status.staged > 0 && <Text color="orange.400">{status.staged} staged</Text>}
          <Text color="gray.500">{status.in_herd} in herd</Text>
        </HStack>
      )}

      {/* Canvas */}
      <Box flex={1}>
        <ISPIsoflow
          initialData={scene}
          width="100%"
          height="100%"
          editorMode="EXPLORABLE_READONLY"
          onItemClick={onSiteSelect}
        />
      </Box>

      {/* Site nav chips — reliable click-through regardless of canvas event support */}
      {sites.length > 0 && (
        <Box px={4} py={2} bg="gray.950" borderTop="1px solid" borderColor="gray.800">
          <HStack spacing={2} flexWrap="wrap">
            <Text fontSize="xs" color="gray.600" mr={1}>Sites:</Text>
            {sites.map(s => (
              <Button
                key={s.id}
                size="xs"
                variant="outline"
                colorScheme={STATUS_COLOR[s.fleet_status] ?? 'blue'}
                onClick={() => onSiteSelect(s.id)}
                letterSpacing="0.05em"
              >
                {s.name}
                <Badge ml={1} fontSize="0.6em" colorScheme="gray">{s.node_count}</Badge>
              </Button>
            ))}
          </HStack>
        </Box>
      )}
    </Box>
  )
}
