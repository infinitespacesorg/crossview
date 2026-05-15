import { useEffect, useState } from 'react'
import { Box, Button, Text, HStack, Checkbox } from '@chakra-ui/react'
import { ISPIsoflow } from 'isp-fossflow'
import { fetchNodes } from './api.js'
import { buildSiteScene } from './scene-builder.js'

export function SiteView({ siteName, env, onBack }) {
  const [scene, setScene] = useState(null)
  const [nodes, setNodes] = useState([])
  const [error, setError] = useState(null)
  const [penOpen, setPenOpen] = useState(false)

  useEffect(() => {
    fetchNodes(siteName)
      .then(n => { setNodes(n); setScene(buildSiteScene(siteName, n, env)) })
      .catch(e => setError(e.message))
  }, [siteName, env])

  if (error) return <Box p={4} color="red.400">Error: {error}</Box>
  if (!scene) return <Box p={4} color="blue.300">Loading {siteName}…</Box>

  const staged = nodes.filter(n => n.phase === 'Provisioning')

  return (
    <Box display="flex" flexDirection="column" width="100%" height="100%">
      <HStack px={4} py={2} bg="gray.900" borderBottom="1px solid" borderColor="gray.700" spacing={4}>
        <Button size="xs" variant="outline" colorScheme="blue" onClick={onBack}>← Range</Button>
        <Text fontWeight={600} color="blue.300" letterSpacing="0.08em" fontSize="sm">{siteName.toUpperCase()}</Text>
        <Text color="gray.500" fontSize="xs">{env}</Text>
        <Text fontSize="xs">{nodes.length} nodes</Text>
        {staged.length > 0 && (
          <Button size="xs" colorScheme="orange" variant="outline" ml="auto" onClick={() => setPenOpen(true)}>
            Pen: {staged.length} staged
          </Button>
        )}
      </HStack>
      <Box flex={1} position="relative">
        <ISPIsoflow initialData={scene} width="100%" height="100%" editorMode="EDITABLE" />
        {penOpen && (
          <PenPanel nodes={staged} onClose={() => setPenOpen(false)} onMuster={() => setPenOpen(false)} />
        )}
      </Box>
    </Box>
  )
}

function PenPanel({ nodes, onClose, onMuster }) {
  const [selected, setSelected] = useState(new Set())
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState(null)

  const toggle = (id) =>
    setSelected(s => { const n = new Set(s); n.has(id) ? n.delete(id) : n.add(id); return n })

  const doMuster = async () => {
    setLoading(true)
    try {
      const { musterNodes } = await import('./api.js')
      const r = await musterNodes([...selected])
      setResult(r)
      setTimeout(onMuster, 1500)
    } catch (e) {
      setResult({ error: e.message })
    } finally {
      setLoading(false)
    }
  }

  return (
    <Box
      position="absolute" right={0} top={0} bottom={0} width="280px"
      bg="gray.900" borderLeft="1px solid" borderColor="gray.700"
      p={4} display="flex" flexDirection="column" gap={2} zIndex={10}
    >
      <HStack justify="space-between" mb={2}>
        <Text color="orange.300" fontWeight={600} fontSize="xs">PEN — Staged Nodes</Text>
        <Button size="xs" variant="ghost" onClick={onClose}>✕</Button>
      </HStack>
      {nodes.map(n => (
        <HStack key={n.id} spacing={2} fontSize="xs">
          <Checkbox isChecked={selected.has(n.id)} onChange={() => toggle(n.id)} colorScheme="blue" size="sm" />
          <Text color="gray.200">{n.name}</Text>
          <Text color="gray.500" ml="auto">{n.role}</Text>
        </HStack>
      ))}
      {result && (
        <Text fontSize="xs" color={result.error ? 'red.400' : 'green.400'} mt={2}>
          {result.error ?? `Mustered ${result.approved?.length ?? 0}`}
        </Text>
      )}
      <Button mt="auto" size="sm" colorScheme="blue" isDisabled={selected.size === 0 || loading} isLoading={loading} onClick={doMuster}>
        Muster {selected.size} nodes
      </Button>
    </Box>
  )
}
