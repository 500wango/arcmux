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
import { ArrowRight, Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'

import { HeroTerminalDemo } from '../hero-terminal-demo'

interface HeroProps {
  className?: string
  isAuthenticated?: boolean
}

export function Hero(props: HeroProps) {
  const { t } = useTranslation()

  return (
    <section className='relative z-10 overflow-hidden px-6 pt-32 pb-20 md:pt-44 md:pb-32 lg:pt-52 lg:pb-36'>
      {/* Sleek ambient background grid with radial mask */}
      <div
        aria-hidden
        className='tech-grid pointer-events-none absolute inset-0 -z-10 [mask-image:radial-gradient(ellipse_60%_50%_at_50%_25%,#000_60%,transparent_100%)] opacity-30 dark:opacity-20'
      />

      {/* Luminous ambient aurora glow */}
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 -z-10 opacity-30 dark:opacity-20'
        style={{
          background: [
            'radial-gradient(ellipse 65% 35% at 50% 12%, color-mix(in oklch, var(--primary) 35%, transparent) 0%, transparent 70%)',
            'radial-gradient(ellipse 45% 30% at 80% 25%, color-mix(in oklch, var(--chart-2) 25%, transparent) 0%, transparent 70%)',
          ].join(', '),
        }}
      />

      <div className='mx-auto max-w-6xl'>
        <div className='flex flex-col items-center text-center'>
          {/* Top badge with live pulsating dot */}
          <div className='landing-animate-fade-up group mb-6 inline-flex items-center gap-2 rounded-full border border-primary/30 bg-primary/10 px-4 py-1.5 text-xs font-semibold text-primary opacity-0 shadow-xs backdrop-blur-md transition-all duration-300 hover:border-primary/50 hover:bg-primary/15 hover:shadow-sm'>
            <span className='relative flex size-2'>
              <span className='absolute inline-flex size-full animate-ping rounded-full bg-primary opacity-75'></span>
              <span className='relative inline-flex size-2 rounded-full bg-primary'></span>
            </span>
            <span>{t('Enterprise AI Gateway Platform')}</span>
            <Sparkles className='size-3 opacity-80 transition-transform duration-300 group-hover:rotate-12' />
          </div>

          {/* Main heading with crisp high-tech gradient */}
          <h1 className='landing-animate-fade-up text-[clamp(2.5rem,5.5vw,4.25rem)] leading-[1.08] font-extrabold tracking-tight opacity-0'>
            {t('One Gateway for All')}
            <br />
            <span className='bg-gradient-to-r from-primary via-amber-500 to-yellow-500 bg-clip-text text-transparent drop-shadow-xs'>
              {t('Your AI Models')}
            </span>
          </h1>

          {/* Subtitle */}
          <p className='landing-animate-fade-up text-muted-foreground mx-auto mt-6 max-w-2xl text-base leading-relaxed opacity-0 md:text-lg'>
            {t('Aggregate every upstream AI provider behind a single, unified API. Route requests, control costs, and scale without switching endpoints.')}
          </p>

          {/* Action buttons */}
          <div className='landing-animate-fade-up mt-10 flex flex-wrap items-center justify-center gap-3.5 opacity-0'>
            {props.isAuthenticated ? (
              <Button
                className='group relative h-11 rounded-xl px-6 text-sm font-semibold shadow-md transition-all duration-200 hover:shadow-lg hover:shadow-primary/20'
                render={<Link to='/dashboard' />}
              >
                {t('Go to Dashboard')}
                <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-1' />
              </Button>
            ) : (
              <>
                <Button
                  className='group relative h-11 rounded-xl px-6 text-sm font-semibold shadow-md transition-all duration-200 hover:shadow-lg hover:shadow-primary/20'
                  render={<Link to='/sign-up' />}
                >
                  {t('Get Started')}
                  <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-1' />
                </Button>
                <Button
                  variant='outline'
                  className='h-11 rounded-xl border-border/60 bg-background/60 px-5 text-sm font-medium backdrop-blur-md transition-all duration-200 hover:border-border hover:bg-muted/60 hover:shadow-xs'
                  render={<Link to='/pricing' />}
                >
                  {t('View Pricing')}
                </Button>
              </>
            )}
          </div>
        </div>

        {/* Terminal demo */}
        <div className='landing-animate-fade-up mt-16 opacity-0 md:mt-20'>
          <HeroTerminalDemo className='mx-auto max-w-3xl' />
        </div>
      </div>
    </section>
  )
}
