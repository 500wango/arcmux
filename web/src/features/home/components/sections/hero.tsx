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
import { ArrowRight, Check, Copy, KeyRound, Terminal } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

import { MeshRouterVisualizer } from '../mesh-router-visualizer'

interface HeroProps {
  className?: string
  isAuthenticated?: boolean
}

export function Hero(props: HeroProps) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  const quickCurl = 'curl -sS https://api.arcmux.com/v1/models'

  const handleCopy = () => {
    navigator.clipboard.writeText(quickCurl)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <section className={cn('relative z-10 overflow-hidden px-6 pt-32 pb-16 md:pt-40 md:pb-24', props.className)}>
      {/* Sleek precision engineering background grid */}
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 -z-10 [mask-image:radial-gradient(ellipse_60%_50%_at_50%_20%,#000_70%,transparent_100%)] opacity-30 dark:opacity-20'
        style={{
          backgroundImage:
            'linear-gradient(to right, currentColor 1px, transparent 1px), linear-gradient(to bottom, currentColor 1px, transparent 1px)',
          backgroundSize: '32px 32px',
        }}
      />

      <div className='mx-auto max-w-6xl'>
        <div className='flex flex-col items-center text-center'>
          {/* Live Engine Status Pill */}
          <div className='landing-animate-fade-up mb-6 inline-flex items-center gap-2.5 rounded-full border border-border/80 bg-background/80 px-4 py-1.5 font-mono text-xs font-medium text-foreground shadow-xs backdrop-blur-md'>
            <span className='relative flex size-2'>
              <span className='absolute inline-flex size-full animate-ping rounded-full bg-emerald-500 opacity-75' />
              <span className='relative inline-flex size-2 rounded-full bg-emerald-500' />
            </span>
            <span>SYSTEM OPERATIONAL</span>
            <span className='text-muted-foreground/40'>|</span>
            <span className='text-emerald-600 dark:text-emerald-400 font-semibold'>
              40+ PROVIDERS UNIFIED
            </span>
          </div>

          {/* Main Title: Confident & Engineering-focused */}
          <h1 className='landing-animate-fade-up text-[clamp(2.5rem,5.5vw,4.5rem)] leading-[1.06] font-extrabold tracking-tight text-foreground'>
            {t('The High-Performance')}
            <br />
            <span className='bg-gradient-to-r from-primary via-emerald-500 to-teal-400 bg-clip-text text-transparent'>
              {t('AI API Multiplexer & Gateway')}
            </span>
          </h1>

          {/* Subtitle: High-signal, zero fluff */}
          <p className='landing-animate-fade-up text-muted-foreground mx-auto mt-6 max-w-2xl text-base leading-relaxed md:text-lg'>
            {t(
              'Aggregate OpenAI, Claude, Gemini, DeepSeek, and 40+ providers into a single resilient endpoint. Sub-millisecond failover, instant stream dispatch, and 40% standard pricing.'
            )}
          </p>

          {/* Action Row */}
          <div className='landing-animate-fade-up mt-9 flex flex-wrap items-center justify-center gap-3.5'>
            {props.isAuthenticated ? (
              <Button
                size='lg'
                className='group gap-2 rounded-xl px-6 text-sm font-semibold shadow-md transition-all hover:shadow-lg hover:shadow-primary/20'
                render={<Link to='/dashboard' />}
              >
                <KeyRound className='size-4' />
                {t('Go to Dashboard')}
                <ArrowRight className='size-4 transition-transform group-hover:translate-x-1' />
              </Button>
            ) : (
              <>
                <Button
                  size='lg'
                  className='group gap-2 rounded-xl px-6 text-sm font-semibold shadow-md transition-all hover:shadow-lg hover:shadow-primary/20'
                  render={<Link to='/sign-up' />}
                >
                  <KeyRound className='size-4' />
                  {t('Get Free API Key')}
                  <ArrowRight className='size-4 transition-transform group-hover:translate-x-1' />
                </Button>
                <Button
                  size='lg'
                  variant='outline'
                  className='rounded-xl border-border/80 bg-background/60 px-5 text-sm font-medium backdrop-blur-md transition-all hover:border-border hover:bg-muted/60'
                  render={<Link to='/pricing' />}
                >
                  {t('Model Square & Pricing')}
                </Button>
              </>
            )}
          </div>

          {/* One-liner quick test pill */}
          <div className='landing-animate-fade-up mt-6 inline-flex items-center gap-2 rounded-xl border border-border/70 bg-zinc-950/80 px-3.5 py-1.5 font-mono text-xs text-zinc-300 shadow-xs backdrop-blur-md'>
            <Terminal className='size-3.5 text-primary' />
            <span>{quickCurl}</span>
            <button
              type='button'
              onClick={handleCopy}
              className='ml-1.5 text-zinc-400 hover:text-zinc-100'
              title={t('Copy')}
            >
              {copied ? (
                <Check className='size-3.5 text-emerald-400' />
              ) : (
                <Copy className='size-3.5' />
              )}
            </button>
          </div>
        </div>

        {/* Live Interactive Mesh Router Canvas */}
        <div className='landing-animate-fade-up mt-14 md:mt-20'>
          <MeshRouterVisualizer />
        </div>
      </div>
    </section>
  )
}
