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

import { createInstance } from 'i18next'
import { renderToStaticMarkup } from 'react-dom/server'
import { I18nextProvider } from 'react-i18next'

import { UptimeHistory } from '../uptime-panel'

test('heartbeat bars expose each recorded status and time in chronological order', async () => {
  const i18n = createInstance()
  await i18n.init({ lng: 'en', resources: { en: { translation: {} } } })
  const html = renderToStaticMarkup(
    <I18nextProvider i18n={i18n}>
      <UptimeHistory
        heartbeats={[
          { status: 1, time: '2026-09-13 08:28:00' },
          { status: 0, time: '2026-09-13 08:29:00' },
        ]}
      />
    </I18nextProvider>
  )
  const bars = html.match(/<span[^>]+role="img"[^>]*>/g) ?? []
  assert.equal(bars.length, 2)
  assert.match(bars[0], /08:28:00 UTC · Online/)
  assert.match(bars[0], /bg-emerald-500/)
  assert.match(bars[1], /08:29:00 UTC · Failed/)
  assert.match(bars[1], /bg-red-500/)
  assert.equal((html.match(/title="No data"/g) ?? []).length, 58)
})
