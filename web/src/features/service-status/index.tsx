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
import { useTranslation } from 'react-i18next'

import { SectionPageLayout } from '@/components/layout'
import { UptimePanel } from '@/features/dashboard/components/overview/uptime-panel'
import { useStatus } from '@/hooks/use-status'

export function ServiceStatus() {
  const { t } = useTranslation()
  const { status, loading } = useStatus()

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Service Status')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        {loading && (
          <p className='text-muted-foreground text-sm'>{t('Loading...')}</p>
        )}
        {!loading && status && status.uptime_kuma_enabled !== false && (
          <UptimePanel />
        )}
        {!loading && (!status || status.uptime_kuma_enabled === false) && (
          <p className='text-muted-foreground text-sm'>
            {t('No uptime monitoring configured')}
          </p>
        )}
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
