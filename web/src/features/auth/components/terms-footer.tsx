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
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

import type { SystemStatus } from '../types'

interface TermsFooterProps {
  variant?: 'sign-in' | 'sign-up'
  className?: string
  status?: SystemStatus | null
}

export function TermsFooter({
  variant = 'sign-in',
  className,
}: TermsFooterProps) {
  const { t } = useTranslation()
  const text =
    variant === 'sign-in'
      ? t('By clicking sign in, you agree to our')
      : t('By creating an account, you agree to our')

  return (
    <p className={cn('text-muted-foreground text-center text-xs', className)}>
      {text}{' '}
      <Link
        to='/terms'
        className='hover:text-primary underline underline-offset-4'
      >
        {t('Terms & Conditions')}
      </Link>{' '}
      {t('and')}{' '}
      <Link
        to='/privacy'
        className='hover:text-primary underline underline-offset-4'
      >
        {t('Privacy Policy')}
      </Link>
      .
    </p>
  )
}
