/*
Copyright (C) 2023-2026 QuantumNous

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

For commercial licensing, please contact support@quantumnous.com
*/
import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import {
  parseSidebarModulesAdmin,
  serializeSidebarModulesAdmin,
} from '../config'

describe('sidebar module compatibility', () => {
  test('drops legacy chat controls while keeping Playground available', () => {
    const parsed = parseSidebarModulesAdmin(
      JSON.stringify({
        chat: { enabled: true, playground: true, chat: true },
        console: { enabled: true, detail: true },
      })
    )

    assert.equal(parsed.chat, undefined)
    assert.equal(parsed.console.playground, true)
    const serialized = JSON.parse(serializeSidebarModulesAdmin(parsed))
    assert.equal(serialized.chat, undefined)
  })
})
