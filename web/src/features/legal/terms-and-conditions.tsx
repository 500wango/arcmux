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

import { useSystemConfig } from '@/hooks/use-system-config'

import { getUserAgreement } from './api'
import { getDefaultTermsAndConditions } from './defaults'
import { LegalDocument } from './legal-document'

export function TermsAndConditions() {
  const { t, i18n } = useTranslation()
  const { systemName } = useSystemConfig()
  const displayName = systemName || 'ArcMux'

  return (
    <LegalDocument
      title={t('Terms & Conditions')}
      subtitle={t('Terms of Service and Acceptable Use Policy')}
      badge={t('Legal & Compliance')}
      queryKey='terms-and-conditions'
      fetchDocument={getUserAgreement}
      defaultContent={getDefaultTermsAndConditions(displayName, i18n.language)}
      emptyMessage={t(
        'The administrator has not configured a user agreement yet.'
      )}
    />
  )
}
