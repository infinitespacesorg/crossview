import { useState } from 'react'
import { Box, HStack, Button, Text } from '@chakra-ui/react'
import { RangeView } from './RangeView.jsx'
import { SiteView } from './SiteView.jsx'

const ENV_BG = { dev: '#4b4237', staging: '#1e3450', prod: '#3a1010' }

export function WorldView() {
  const [env, setEnv] = useState('staging')
  const [view, setView] = useState('range')
  const [selectedSite, setSelectedSite] = useState(null)

  const handleSiteSelect = (siteId) => {
    setSelectedSite(siteId)
    setView('site')
  }

  return (
    <Box display="flex" flexDirection="column" width="100%" height="calc(100vh - 64px)" bg="#1a1a1a" color="gray.200">
      <HStack px={5} py={2} bg="gray.900" borderBottom="1px solid" borderColor="gray.700" spacing={6} flexShrink={0}>
        <Text fontSize="sm" fontWeight={600} letterSpacing="0.1em" color="blue.300">CROSSVIEW</Text>
        <Text color="gray.500" fontSize="xs">World View</Text>
        <HStack spacing={2} ml={4}>
          {['dev', 'staging', 'prod'].map(e => (
            <Button
              key={e}
              size="xs"
              variant={env === e ? 'solid' : 'outline'}
              colorScheme="blue"
              bg={env === e ? ENV_BG[e] : 'transparent'}
              onClick={() => { setEnv(e); setView('range'); setSelectedSite(null) }}
              textTransform="uppercase"
              letterSpacing="0.05em"
            >
              {e}
            </Button>
          ))}
        </HStack>
      </HStack>

      <Box flex={1} position="relative" overflow="hidden">
        {view === 'range' && <RangeView env={env} onSiteSelect={handleSiteSelect} />}
        {view === 'site' && selectedSite && (
          <SiteView siteName={selectedSite} env={env} onBack={() => setView('range')} />
        )}
      </Box>
    </Box>
  )
}
