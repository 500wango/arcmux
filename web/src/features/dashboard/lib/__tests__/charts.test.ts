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
import { describe, test } from 'node:test'

import type { QuotaDataItem } from '../../types'
import { processUserChartData } from '../charts'

function localTimestamp(day: number, hour = 12): number {
  return Math.floor(new Date(2026, 7, day, hour).getTime() / 1000)
}

describe('user dashboard chart data', () => {
  test('ranks and trends by token usage when the token metric is selected', () => {
    const data: QuotaDataItem[] = [
      {
        username: 'alice',
        created_at: localTimestamp(15),
        quota: 900,
        token_used: 10,
        count: 1,
      },
      {
        username: 'alice',
        created_at: localTimestamp(17),
        quota: 900,
        token_used: 100,
        count: 2,
      },
      {
        username: 'bob',
        created_at: localTimestamp(17),
        quota: 50,
        token_used: 20,
        count: 1,
      },
    ]

    const result = processUserChartData(data, 'day', undefined, 10, 'tokens', {
      start_timestamp: localTimestamp(15, 8),
      end_timestamp: localTimestamp(17, 18),
    })

    const rankValues = result.spec_user_rank.data[0].values as Array<
      Record<string, unknown>
    >
    assert.deepEqual(
      rankValues.map((item) => [item.User, item.rawValue]),
      [
        ['alice', 110],
        ['bob', 20],
      ]
    )

    const trendValues = result.spec_user_trend.data[0].values as Array<
      Record<string, unknown>
    >
    const aliceValues = trendValues
      .filter((item) => item.User === 'alice')
      .map((item) => item.rawValue)
    assert.deepEqual(aliceValues, [10, 0, 100])
  })

  test('keeps all 14 daily buckets when traffic exists on only five days', () => {
    const data: QuotaDataItem[] = [14, 15, 16, 17, 18].map((day) => ({
      username: 'alice',
      created_at: localTimestamp(day),
      quota: 100,
      token_used: day,
      count: 1,
    }))

    const result = processUserChartData(data, 'day', undefined, 10, 'tokens', {
      start_timestamp: localTimestamp(4),
      end_timestamp: localTimestamp(18),
    })

    const trendValues = result.spec_user_trend.data[0].values as Array<
      Record<string, unknown>
    >
    assert.equal(trendValues.length, 14)
    assert.equal(trendValues.filter((item) => item.rawValue === 0).length, 9)
  })
})
