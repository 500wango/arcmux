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
import { CheckCircle2, ShieldCheck, Zap } from 'lucide-react'
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
      <div className='relative hidden flex-col justify-between overflow-hidden border-r border-border/70 bg-zinc-950 p-10 text-zinc-100 md:flex'>
        {/* Subtle engineering grid */}
        <div
          aria-hidden
          className='pointer-events-none absolute inset-0 opacity-10'
          style={{
            backgroundImage:
              'linear-gradient(to right, currentColor 1px, transparent 1px), linear-gradient(to bottom, currentColor 1px, transparent 1px)',
            backgroundSize: '28px 28px',
          }}
        />

        {/* Top brand header */}
        <div className='relative z-10'>
          <Link
            to='/'
            className='flex items-center gap-3.5 transition-opacity hover:opacity-85'
          >
            <div className='relative size-10 overflow-hidden rounded-xl border border-zinc-800 bg-zinc-900 shadow-xs'>
              {loading ? (
                <Skeleton className='size-full' />
              ) : (
                <img
                  src={logo || '/logo-icon.svg'}
                  alt={t('Logo')}
                  className='size-full object-contain'
                />
              )}
            </div>
            {loading ? (
              <Skeleton className='h-6 w-24' />
            ) : (
              <span className='font-mono text-lg font-bold tracking-tight text-white'>
                {systemName || 'ArcMux'}
              </span>
            )}
          </Link>
        </div>

        {/* Center content */}
        <div className='relative z-10 my-auto max-w-md py-12'>
          <div className='mb-6 inline-flex items-center gap-2 rounded-full border border-emerald-500/30 bg-emerald-500/10 px-3 py-1 font-mono text-xs font-medium text-emerald-400'>
            <span className='size-1.5 animate-pulse rounded-full bg-emerald-400' />
            <span>{t('GATEWAY ENGINE OPERATIONAL')}</span>
          </div>

          <h2 className='text-3xl font-extrabold tracking-tight text-white md:text-4xl'>
            {t('High-Throughput AI Multiplexer')}
          </h2>

          <p className='text-zinc-400 mt-4 text-sm leading-relaxed'>
            {t(
              'Unified AI API gateway. Route, load-balance, and scale across OpenAI, Claude, Gemini, and DeepSeek with sub-millisecond dispatching.'
            )}
          </p>

          {/* Architecture telemetry badges */}
          <div className='mt-8 space-y-3 font-mono text-xs text-zinc-300'>
            <div className='flex items-center gap-2.5 rounded-lg border border-zinc-800/80 bg-zinc-900/50 p-2.5'>
              <ShieldCheck className='size-4 text-emerald-400' />
              <span>{t('Sub-millisecond failover & circuit breaking')}</span>
            </div>
            <div className='flex items-center gap-2.5 rounded-lg border border-zinc-800/80 bg-zinc-900/50 p-2.5'>
              <Zap className='size-4 text-amber-400' />
              <span>{t('Unbuffered SSE stream acceleration')}</span>
            </div>
            <div className='flex items-center gap-2.5 rounded-lg border border-zinc-800/80 bg-zinc-900/50 p-2.5'>
              <CheckCircle2 className='size-4 text-primary' />
              <span>{t('Zero SDK refactoring (OpenAI & Claude compatible)')}</span>
            </div>
          </div>
        </div>

        {/* Bottom copyright & spec */}
        <div className='relative z-10 flex items-center justify-between border-t border-zinc-800/80 pt-4 font-mono text-xs text-zinc-500'>
          <span>&copy; {new Date().getFullYear()} {systemName || 'ArcMux'}</span>
          <span>v1.0.0-PROD</span>
        </div>
      </div>

      {/* Form panel */}
      <div className='relative flex items-center justify-center px-4'>
        {/* Mobile logo */}
        <Link
          to='/'
          className='absolute top-4 left-4 z-10 flex items-center gap-2.5 transition-opacity hover:opacity-80 md:hidden'
        >
          <div className='relative size-9 overflow-hidden rounded-xl border border-border/70 bg-card/60'>
            {loading ? (
              <Skeleton className='size-full' />
            ) : (
              <img
                src={logo || '/logo-icon.svg'}
                alt={t('Logo')}
                className='size-full object-contain'
              />
            )}
          </div>
          {loading ? (
            <Skeleton className='h-6 w-24' />
          ) : (
            <h1 className='font-mono text-base font-bold'>{systemName || 'ArcMux'}</h1>
          )}
        </Link>

        <div className='w-full max-w-sm py-8'>
          {children}
        </div>
      </div>
    </div>
  )
}