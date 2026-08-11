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
import { ArrowRight, KeyRound, Terminal } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'
import { Button } from '@/components/ui/button'

interface CTAProps {
  className?: string
  isAuthenticated?: boolean
}

export function CTA(props: CTAProps) {
  const { t } = useTranslation()

  if (props.isAuthenticated) {
    return null
  }

  return (
    <section className='relative z-10 overflow-hidden px-6 py-20 md:py-28'>
      <div className='mx-auto max-w-6xl'>
        <AnimateInView
          className='relative overflow-hidden rounded-3xl border border-border/80 bg-zinc-950/90 p-8 text-center text-zinc-100 shadow-2xl backdrop-blur-xl md:p-14'
          animation='scale-in'
        >
          {/* Subtle micro-grid accent */}
          <div
            aria-hidden
            className='pointer-events-none absolute inset-0 opacity-5'
            style={{
              backgroundImage:
                'radial-gradient(circle, currentColor 1px, transparent 1px)',
              backgroundSize: '20px 20px',
            }}
          />

          <div className='relative z-10 mx-auto max-w-2xl'>
            <div className='mb-4 inline-flex items-center gap-2 rounded-full border border-emerald-500/30 bg-emerald-500/10 px-3 py-1 font-mono text-xs font-semibold text-emerald-400'>
              <Terminal className='size-3.5' />
              <span>{t('Instant Integration in 60s')}</span>
            </div>

            <h2 className='text-3xl font-extrabold tracking-tight md:text-5xl'>
              {t('Ready to scale with zero vendor lock-in?')}
            </h2>

            <p className='text-zinc-400 mt-4 text-sm leading-relaxed md:text-base'>
              {t(
                'Generate your API token now to start routing OpenAI, Claude, Gemini, and DeepSeek requests with unified billing and 40% standard pricing.'
              )}
            </p>

            <div className='mt-8 flex flex-wrap items-center justify-center gap-3.5'>
              <Button
                size='lg'
                className='group gap-2 rounded-xl px-6 text-sm font-semibold shadow-md transition-all hover:shadow-lg hover:shadow-primary/20'
                render={<Link to='/sign-up' />}
              >
                <KeyRound className='size-4' />
                {t('Get Started Free')}
                <ArrowRight className='size-4 transition-transform group-hover:translate-x-1' />
              </Button>
              <Button
                size='lg'
                variant='outline'
                className='rounded-xl border-zinc-800 bg-zinc-900/80 px-5 text-sm font-medium text-zinc-200 hover:bg-zinc-800 hover:text-white'
                render={<Link to='/pricing' />}
              >
                {t('Explore Models & Rates')}
              </Button>
            </div>
          </div>
        </AnimateInView>
      </div>
    </section>
  )
}