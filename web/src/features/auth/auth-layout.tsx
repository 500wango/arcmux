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
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'

type AuthLayoutProps = {
  children: React.ReactNode
}

export function AuthLayout({ children }: AuthLayoutProps) {
  const { t } = useTranslation()
  const { systemName, logo, loading } = useSystemConfig()

  return (
    <div className='relative grid min-h-svh max-w-none md:grid-cols-2'>
      {/* Brand panel — visible on md+ */}
      <div className='bg-muted/30 relative hidden flex-col justify-between p-10 md:flex'>
        <Link
          to='/'
          className='flex items-center gap-2 transition-opacity hover:opacity-80'
        >
          <div className='relative h-8 w-8'>
            {loading ? (
              <Skeleton className='absolute inset-0 rounded-full' />
            ) : (
              <img
                src={logo}
                alt={t('Logo')}
                className='h-8 w-8 rounded-full object-cover'
              />
            )}
          </div>
          {loading ? (
            <Skeleton className='h-6 w-24' />
          ) : (
            <h1 className='text-xl font-medium'>{systemName}</h1>
          )}
        </Link>

        <div className='flex-1 flex flex-col justify-center px-8'>
          <div className='mx-auto max-w-sm'>
            <div className='text-primary mb-6 flex size-16 items-center justify-center rounded-2xl border border-primary/20 bg-primary/5'>
              <svg
                className='size-8'
                viewBox='0 0 24 24'
                fill='none'
                stroke='currentColor'
                strokeWidth={1.5}
                strokeLinecap='round'
                strokeLinejoin='round'
              >
                <path d='M12 2L2 7l10 5 10-5-10-5z' />
                <path d='M2 17l10 5 10-5' />
                <path d='M2 12l10 5 10-5' />
              </svg>
            </div>
            <h2 className='text-2xl font-bold tracking-tight'>
              {systemName}
            </h2>
            <p className='text-muted-foreground mt-3 text-sm leading-relaxed'>
              {t(
                'Unified AI API gateway. Manage, monitor, and route API requests across multiple AI providers.'
              )}
            </p>
          </div>
        </div>

        <p className='text-muted-foreground/50 text-xs'>
          &copy; {new Date().getFullYear()} {systemName}
        </p>
      </div>

      {/* Form panel */}
      <div className='flex items-center justify-center px-4'>
        {/* Mobile logo */}
        <Link
          to='/'
          className='absolute top-4 left-4 z-10 flex items-center gap-2 transition-opacity hover:opacity-80 md:hidden'
        >
          <div className='relative h-8 w-8'>
            {loading ? (
              <Skeleton className='absolute inset-0 rounded-full' />
            ) : (
              <img
                src={logo}
                alt={t('Logo')}
                className='h-8 w-8 rounded-full object-cover'
              />
            )}
          </div>
          {loading ? (
            <Skeleton className='h-6 w-24' />
          ) : (
            <h1 className='text-xl font-medium'>{systemName}</h1>
          )}
        </Link>

        <div className='w-full max-w-sm py-8'>
          {children}
        </div>
      </div>
    </div>
  )
}