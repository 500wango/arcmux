/*
Copyright (C) 2023-2026 ArcMux contributors

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
import assert from 'node:assert/strict'
import { test } from 'node:test'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createInstance } from 'i18next'
import { renderToStaticMarkup } from 'react-dom/server'
import { I18nextProvider } from 'react-i18next'

import { useSidebarData } from '@/hooks/use-sidebar-data'

import { ServiceStatus } from '..'

function StatusNavigation() {
  const { navGroups } = useSidebarData()
  const item = navGroups
    .flatMap((group) => group.items)
    .find((item) => item.url === '/service-status')
  return item ? <a href={item.url}>{item.title}</a> : null
}

async function renderStatus(enabled: boolean) {
  const client = new QueryClient()
  client.setQueryData(['status'], { uptime_kuma_enabled: enabled })
  const i18n = createInstance()
  await i18n.init({ lng: 'en', resources: { en: { translation: {} } } })
  try {
    return renderToStaticMarkup(
      <I18nextProvider i18n={i18n}>
        <QueryClientProvider client={client}>
          <StatusNavigation />
          <ServiceStatus />
        </QueryClientProvider>
      </I18nextProvider>
    )
  } finally {
    client.clear()
  }
}

test('enabled Uptime Kuma exposes the sidebar destination and monitoring panel without an admin role', async () => {
  const html = await renderStatus(true)
  assert.match(html, /href="\/service-status"/)
  assert.match(html, /Grouped monitor status from Uptime Kuma/)
})

test('disabled Uptime Kuma hides the sidebar entry and monitoring panel', async () => {
  const html = await renderStatus(false)
  assert.doesNotMatch(
    html,
    /href="\/service-status"|Grouped monitor status from Uptime Kuma/
  )
  assert.match(html, /No uptime monitoring configured/)
})
