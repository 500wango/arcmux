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
import {
  createMemoryHistory,
  createRootRoute,
  createRouter,
  RouterContextProvider,
} from '@tanstack/react-router'
import { createInstance } from 'i18next'
import { renderToStaticMarkup } from 'react-dom/server'
import { I18nextProvider } from 'react-i18next'

import type { PricingModel } from '@/features/pricing/types'

import { CodeIntegrationTabs } from '../code-integration-tabs'
import { MeshRouterVisualizer } from '../mesh-router-visualizer'
import { ProviderMatrix } from '../provider-matrix'

async function renderCatalog(models: PricingModel[]) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  client.setQueryData(['status'], {})
  client.setQueryData(['pricing'], {
    success: true,
    data: models,
    vendors: [{ id: 7, name: 'Configured vendor' }],
    group_ratio: {},
    usable_group: {},
    supported_endpoint: {},
    auto_groups: [],
  })
  const i18n = createInstance()
  await i18n.init({
    lng: 'en',
    resources: { en: { translation: {} } },
    interpolation: { escapeValue: false },
  })
  const router = createRouter({
    routeTree: createRootRoute(),
    history: createMemoryHistory(),
  })
  try {
    return renderToStaticMarkup(
      <I18nextProvider i18n={i18n}>
        <QueryClientProvider client={client}>
          <RouterContextProvider router={router}>
            <ProviderMatrix />
            <CodeIntegrationTabs />
            <MeshRouterVisualizer />
          </RouterContextProvider>
        </QueryClientProvider>
      </I18nextProvider>
    )
  } finally {
    client.clear()
  }
}

test('homepage uses catalog model names and selects a Chat Completions model for the code example', async () => {
  const base = {
    id: 1,
    vendor_id: 7,
    quota_type: 0,
    model_ratio: 1,
    completion_ratio: 1,
    enable_groups: ['default'],
  }
  const html = await renderCatalog([
    {
      ...base,
      model_name: 'configured-image',
      supported_endpoint_types: ['image-generation'],
    },
    {
      ...base,
      id: 2,
      model_name: 'configured-chat',
      supported_endpoint_types: ['openai'],
    },
  ])
  assert.match(html, /configured-image/)
  assert.match(html, /Configured vendor/)
  const code = html
    .match(/<code[^>]*>([\s\S]*?)<\/code>/g)
    ?.find((block) => block.includes('curl -X POST'))
  assert.ok(code, 'the page exposes a copyable request example')
  assert.match(code, /configured-chat/)
  assert.doesNotMatch(code, /configured-image/)
  assert.doesNotMatch(html, /gpt-4o|claude-3-7-sonnet|deepseek-r1|Azure EastUS/)
})

test('an empty catalog shows empty states without inventing fallback models or a request example', async () => {
  const html = await renderCatalog([])
  assert.match(html, /No models available/)
  assert.match(
    html,
    /No model with Chat Completions support is currently listed/
  )
  assert.doesNotMatch(html, /curl -X POST|gpt-4o|claude-3-7-sonnet|deepseek-r1/)
})
