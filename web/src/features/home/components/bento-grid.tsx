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
import { ArrowRightLeft, Layers, Network, ShieldCheck, Zap } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

export function BentoGrid({ className }: { className?: string }) {
  const { t } = useTranslation()

  return (
    <section className={cn('relative z-10 px-6 py-20 md:py-28', className)}>
      <div className='mx-auto max-w-6xl'>
        {/* Section Header */}
        <div className='mb-14 text-center'>
          <div className='border-primary/20 bg-primary/5 text-primary mb-3 inline-flex items-center gap-1.5 rounded-full border px-3 py-1 font-mono text-xs font-semibold'>
            <Layers className='size-3.5' />
            {t('Engine Architecture')}
          </div>
          <h2 className='text-3xl font-extrabold tracking-tight md:text-4xl'>
            {t('Routing, access, and usage in one gateway')}
          </h2>
          <p className='text-muted-foreground mx-auto mt-3 max-w-2xl text-sm leading-relaxed md:text-base'>
            {t(
              'Manage channel routing, API access, and usage through a Go-based gateway.'
            )}
          </p>
        </div>

        {/* Bento Cards Grid */}
        <div className='grid gap-5 md:grid-cols-12'>
          {/* Card 1: Intelligent Circuit Breaking (Col 7) */}
          <div className='group border-border/80 bg-card/40 hover:border-primary/50 hover:bg-card/70 relative flex flex-col justify-between overflow-hidden rounded-2xl border p-6 shadow-xs backdrop-blur-sm transition-all duration-300 md:col-span-7'>
            <div className='flex items-start justify-between'>
              <div className='border-primary/30 bg-primary/10 text-primary flex size-10 items-center justify-center rounded-xl border'>
                <ShieldCheck className='size-5' />
              </div>
              <span className='rounded-md border border-emerald-500/30 bg-emerald-500/10 px-2 py-0.5 font-mono text-[10px] font-bold text-emerald-500 uppercase'>
                {t('Retry Settings')}
              </span>
            </div>

            <div className='mt-6'>
              <h3 className='text-foreground text-lg font-bold tracking-tight'>
                {t('Configurable channel retries')}
              </h3>
              <p className='text-muted-foreground mt-2 text-xs leading-relaxed md:text-sm'>
                {t(
                  'When retries are enabled and an eligible request fails, the gateway can try another available channel.'
                )}
              </p>
            </div>
          </div>

          {/* Card 2: Universal Protocol Adapter (Col 5) */}
          <div className='group border-border/80 bg-card/40 hover:border-primary/50 hover:bg-card/70 relative flex flex-col justify-between overflow-hidden rounded-2xl border p-6 shadow-xs backdrop-blur-sm transition-all duration-300 md:col-span-5'>
            <div className='flex items-start justify-between'>
              <div className='border-primary/30 bg-primary/10 text-primary flex size-10 items-center justify-center rounded-xl border'>
                <ArrowRightLeft className='size-5' />
              </div>
              <span className='border-primary/30 bg-primary/10 text-primary rounded-md border px-2 py-0.5 font-mono text-[10px] font-bold uppercase'>
                {t('Protocols')}
              </span>
            </div>

            <div className='mt-6'>
              <h3 className='text-foreground text-lg font-bold tracking-tight'>
                {t('Supported API formats')}
              </h3>
              <p className='text-muted-foreground mt-2 text-xs leading-relaxed md:text-sm'>
                {t(
                  'Use OpenAI, Claude, Gemini, and Responses API formats with compatible models and configured channels.'
                )}
              </p>
            </div>

            <div className='mt-6 flex flex-wrap gap-2'>
              {[
                '/v1/chat/completions',
                '/v1/messages',
                '/v1/responses',
                '/v1beta',
              ].map((p) => (
                <span
                  key={p}
                  className='border-border/60 bg-muted/40 text-muted-foreground rounded-md border px-2.5 py-1 font-mono text-xs'
                >
                  {p}
                </span>
              ))}
            </div>
          </div>

          {/* Card 3: Tiered Routing & Group Multipliers (Col 5) */}
          <div className='group border-border/80 bg-card/40 hover:border-primary/50 hover:bg-card/70 relative flex flex-col justify-between overflow-hidden rounded-2xl border p-6 shadow-xs backdrop-blur-sm transition-all duration-300 md:col-span-5'>
            <div className='flex items-start justify-between'>
              <div className='border-primary/30 bg-primary/10 text-primary flex size-10 items-center justify-center rounded-xl border'>
                <Network className='size-5' />
              </div>
              <span className='rounded-md border border-amber-500/30 bg-amber-500/10 px-2 py-0.5 font-mono text-[10px] font-bold text-amber-500 uppercase'>
                {t('Model Pricing')}
              </span>
            </div>

            <div className='mt-6'>
              <h3 className='text-foreground text-lg font-bold tracking-tight'>
                {t('Granular Channel Grouping & Dynamic Rates')}
              </h3>
              <p className='text-muted-foreground mt-2 text-xs leading-relaxed md:text-sm'>
                {t(
                  'Configure model prices and user group multipliers. Check the pricing page for this service’s available models and rates.'
                )}
              </p>
            </div>
          </div>

          {/* Card 4: Streaming responses (Col 7) */}
          <div className='group border-border/80 bg-card/40 hover:border-primary/50 hover:bg-card/70 relative flex flex-col justify-between overflow-hidden rounded-2xl border p-6 shadow-xs backdrop-blur-sm transition-all duration-300 md:col-span-7'>
            <div className='flex items-start justify-between'>
              <div className='border-primary/30 bg-primary/10 text-primary flex size-10 items-center justify-center rounded-xl border'>
                <Zap className='size-5' />
              </div>
              <span className='border-primary/30 bg-primary/10 text-primary rounded-md border px-2 py-0.5 font-mono text-[10px] font-bold uppercase'>
                {t('Streaming')}
              </span>
            </div>

            <div className='mt-6'>
              <h3 className='text-foreground text-lg font-bold tracking-tight'>
                {t('Streaming responses')}
              </h3>
              <p className='text-muted-foreground mt-2 text-xs leading-relaxed md:text-sm'>
                {t(
                  'Forward streaming responses from supported upstream APIs. Response times depend on the model, provider, and network.'
                )}
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
