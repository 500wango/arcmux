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
  CHANNEL_FORM_DEFAULT_VALUES,
  transformFormDataToCreatePayload,
} from '../channel-form'

describe('channel authentication mode', () => {
  test('OAuth creation uses a single empty-key channel', () => {
    const payload = transformFormDataToCreatePayload({
      ...CHANNEL_FORM_DEFAULT_VALUES,
      name: 'Claude OAuth',
      type: 14,
      auth_mode: 'oauth',
      key: 'must-not-be-sent',
      multi_key_mode: 'batch',
      models: 'claude-sonnet-4-5',
    })

    assert.equal(payload.auth_mode, 'oauth')
    assert.equal(payload.mode, 'single')
    assert.equal(payload.channel.key, null)
  })

  test('API Key creation preserves batch mode and credentials', () => {
    const payload = transformFormDataToCreatePayload({
      ...CHANNEL_FORM_DEFAULT_VALUES,
      name: 'Claude API Key',
      type: 14,
      auth_mode: 'api_key',
      key: 'key-one\nkey-two',
      multi_key_mode: 'batch',
      models: 'claude-sonnet-4-5',
    })

    assert.equal(payload.auth_mode, 'api_key')
    assert.equal(payload.mode, 'batch')
    assert.equal(payload.channel.key, 'key-one\nkey-two')
  })
})
