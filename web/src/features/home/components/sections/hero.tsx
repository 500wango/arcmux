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
    <section
      className={cn(
        'relative z-10 px-6 pt-24 pb-16 md:pt-32 md:pb-24',
        props.className
      )}
    >
      <div className='mx-auto max-w-6xl'>
        <div className='grid items-center gap-12 lg:grid-cols-[0.9fr_1.1fr] lg:gap-16'>
          <div>
            {/* Product category */}
            <div className='landing-animate-fade-up border-border bg-muted/40 text-muted-foreground mb-6 inline-flex items-center gap-2 border px-3 py-1.5 font-mono text-[11px] font-medium'>
              <span className='font-semibold text-emerald-600 dark:text-emerald-400'>
                {t('Unified Upstream Ecosystem')}
              </span>
            </div>

            {/* Main Title: Confident & Engineering-focused */}
            <h1 className='landing-animate-fade-up text-foreground max-w-xl text-4xl leading-[1.08] font-semibold tracking-tight md:text-6xl'>
              <span className='text-primary'>
                {t('AI API Multiplexer & Gateway')}
              </span>
            </h1>

            {/* Subtitle: High-signal, zero fluff */}
            <p className='landing-animate-fade-up text-muted-foreground mt-6 max-w-xl text-base leading-relaxed md:text-lg'>
              {t(
                'Access configured AI models through one API, with streaming support, usage tracking, and model-based pricing.'
              )}
            </p>

            {/* Action Row */}
            <div className='landing-animate-fade-up mt-9 flex flex-wrap gap-3'>
              {props.isAuthenticated ? (
                <Button
                  size='lg'
                  className='group gap-2 px-5 text-sm font-semibold'
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
                    className='group gap-2 px-5 text-sm font-semibold'
                    render={<Link to='/sign-up' />}
                  >
                    <KeyRound className='size-4' />
                    {t('Get Started')}
                    <ArrowRight className='size-4 transition-transform group-hover:translate-x-1' />
                  </Button>
                  <Button
                    size='lg'
                    variant='outline'
                    className='border-border px-5 text-sm font-medium'
                    render={<Link to='/pricing' />}
                  >
                    {t('Model Square & Pricing')}
                  </Button>
                </>
              )}
            </div>

            {/* One-liner quick test pill */}
            <div className='landing-animate-fade-up border-border bg-muted/40 text-muted-foreground mt-6 inline-flex max-w-full items-center gap-2 border px-3.5 py-2 font-mono text-xs'>
              <Terminal className='text-primary size-3.5' />
              <span className='truncate'>{quickCurl}</span>
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
          <div className='landing-animate-fade-up'>
            <MeshRouterVisualizer />
          </div>
        </div>
      </div>
    </section>
  )
}
