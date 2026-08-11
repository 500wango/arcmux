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
import {
  Activity,
  ArrowRightLeft,
  Layers,
  Network,
  ShieldCheck,
  Zap,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

export function BentoGrid({ className }: { className?: string }) {
  const { t } = useTranslation()

  return (
    <section className={cn('relative z-10 px-6 py-20 md:py-28', className)}>
      <div className='mx-auto max-w-6xl'>
        {/* Section Header */}
        <div className='mb-14 text-center'>
          <div className='mb-3 inline-flex items-center gap-1.5 rounded-full border border-primary/20 bg-primary/5 px-3 py-1 font-mono text-xs font-semibold text-primary'>
            <Layers className='size-3.5' />
            {t('Engine Architecture')}
          </div>
          <h2 className='text-3xl font-extrabold tracking-tight md:text-4xl'>
            {t('Engineered for High-Concurrency Production')}
          </h2>
          <p className='text-muted-foreground mx-auto mt-3 max-w-2xl text-sm leading-relaxed md:text-base'>
            {t(
              'A hardened, distributed Go gateway core designed for mission-critical AI workloads with sub-millisecond dispatching.'
            )}
          </p>
        </div>

        {/* Bento Cards Grid */}
        <div className='grid gap-5 md:grid-cols-12'>
          {/* Card 1: Intelligent Circuit Breaking (Col 7) */}
          <div className='group relative flex flex-col justify-between overflow-hidden rounded-2xl border border-border/80 bg-card/40 p-6 shadow-xs backdrop-blur-sm transition-all duration-300 hover:border-primary/50 hover:bg-card/70 md:col-span-7'>
            <div className='flex items-start justify-between'>
              <div className='flex size-10 items-center justify-center rounded-xl border border-primary/30 bg-primary/10 text-primary'>
                <ShieldCheck className='size-5' />
              </div>
              <span className='rounded-md border border-emerald-500/30 bg-emerald-500/10 px-2 py-0.5 font-mono text-[10px] font-bold text-emerald-500 uppercase'>
                &lt; 3ms Failover
              </span>
            </div>

            <div className='mt-6'>
              <h3 className='text-lg font-bold tracking-tight text-foreground'>
                {t('Sub-Millisecond Circuit Breaker & Auto-Failover')}
              </h3>
              <p className='text-muted-foreground mt-2 text-xs leading-relaxed md:text-sm'>
                {t(
                  'Continuous health telemetry probes every upstream provider. When an outage or rate-limit spike (429/503) is detected, traffic is rerouted within milliseconds with zero dropped stream packets.'
                )}
              </p>
            </div>

            {/* Visual Telemetry Mockup */}
            <div className='mt-6 rounded-xl border border-border/60 bg-muted/30 p-3.5 font-mono text-xs'>
              <div className='flex items-center justify-between text-muted-foreground text-[11px]'>
                <span>Upstream Node #1 (OpenAI US)</span>
                <span className='font-bold text-red-400'>503 OUTAGE</span>
              </div>
              <div className='my-2 h-1.5 w-full overflow-hidden rounded-full bg-muted'>
                <div className='h-full w-full bg-red-500/80' />
              </div>
              <div className='flex items-center justify-between text-emerald-500 text-[11px] font-semibold'>
                <span>➜ Bypassed to Node #2 (Azure Standby)</span>
                <span>Active (2.1ms)</span>
              </div>
            </div>
          </div>

          {/* Card 2: Universal Protocol Adapter (Col 5) */}
          <div className='group relative flex flex-col justify-between overflow-hidden rounded-2xl border border-border/80 bg-card/40 p-6 shadow-xs backdrop-blur-sm transition-all duration-300 hover:border-primary/50 hover:bg-card/70 md:col-span-5'>
            <div className='flex items-start justify-between'>
              <div className='flex size-10 items-center justify-center rounded-xl border border-primary/30 bg-primary/10 text-primary'>
                <ArrowRightLeft className='size-5' />
              </div>
              <span className='rounded-md border border-primary/30 bg-primary/10 px-2 py-0.5 font-mono text-[10px] font-bold text-primary uppercase'>
                Multi-Protocol
              </span>
            </div>

            <div className='mt-6'>
              <h3 className='text-lg font-bold tracking-tight text-foreground'>
                {t('Universal Protocol Translation')}
              </h3>
              <p className='text-muted-foreground mt-2 text-xs leading-relaxed md:text-sm'>
                {t(
                  'Use OpenAI Chat SDK, Claude Messages API, Google Gemini native format, or Responses protocol seamlessly without changing your application code.'
                )}
              </p>
            </div>

            <div className='mt-6 flex flex-wrap gap-2'>
              {['/v1/chat/completions', '/v1/messages', '/v1/responses', '/v1beta'].map((p) => (
                <span
                  key={p}
                  className='rounded-md border border-border/60 bg-muted/40 px-2.5 py-1 font-mono text-xs text-muted-foreground'
                >
                  {p}
                </span>
              ))}
            </div>
          </div>

          {/* Card 3: Tiered Routing & Group Multipliers (Col 5) */}
          <div className='group relative flex flex-col justify-between overflow-hidden rounded-2xl border border-border/80 bg-card/40 p-6 shadow-xs backdrop-blur-sm transition-all duration-300 hover:border-primary/50 hover:bg-card/70 md:col-span-5'>
            <div className='flex items-start justify-between'>
              <div className='flex size-10 items-center justify-center rounded-xl border border-primary/30 bg-primary/10 text-primary'>
                <Network className='size-5' />
              </div>
              <span className='rounded-md border border-amber-500/30 bg-amber-500/10 px-2 py-0.5 font-mono text-[10px] font-bold text-amber-500 uppercase'>
                40% Standard Rate
              </span>
            </div>

            <div className='mt-6'>
              <h3 className='text-lg font-bold tracking-tight text-foreground'>
                {t('Granular Channel Grouping & Dynamic Rates')}
              </h3>
              <p className='text-muted-foreground mt-2 text-xs leading-relaxed md:text-sm'>
                {t(
                  'Isolate heavy background tasks from low-latency interactive sessions. Apply custom discount multipliers across user groups.'
                )}
              </p>
            </div>

            <div className='mt-6 grid grid-cols-2 gap-2 text-center font-mono text-xs'>
              <div className='rounded-lg border border-border/50 bg-background/50 p-2.5'>
                <div className='text-muted-foreground text-[10px] uppercase'>Default Tier</div>
                <div className='mt-1 font-bold text-emerald-500'>0.40x Ratio</div>
              </div>
              <div className='rounded-lg border border-border/50 bg-background/50 p-2.5'>
                <div className='text-muted-foreground text-[10px] uppercase'>VIP Tier</div>
                <div className='mt-1 font-bold text-primary'>Dedicated QPS</div>
              </div>
            </div>
          </div>

          {/* Card 4: Unbuffered SSE Stream Acceleration (Col 7) */}
          <div className='group relative flex flex-col justify-between overflow-hidden rounded-2xl border border-border/80 bg-card/40 p-6 shadow-xs backdrop-blur-sm transition-all duration-300 hover:border-primary/50 hover:bg-card/70 md:col-span-7'>
            <div className='flex items-start justify-between'>
              <div className='flex size-10 items-center justify-center rounded-xl border border-primary/30 bg-primary/10 text-primary'>
                <Zap className='size-5' />
              </div>
              <span className='rounded-md border border-primary/30 bg-primary/10 px-2 py-0.5 font-mono text-[10px] font-bold text-primary uppercase'>
                Unbuffered SSE
              </span>
            </div>

            <div className='mt-6'>
              <h3 className='text-lg font-bold tracking-tight text-foreground'>
                {t('Real-Time Unbuffered Streaming & TCP Keep-Alive')}
              </h3>
              <p className='text-muted-foreground mt-2 text-xs leading-relaxed md:text-sm'>
                {t(
                  'Optimized streaming pipeline delivers immediate First Token Time (TTFT) with zero gateway buffer delay, combined with intelligent TCP connection pooling.'
                )}
              </p>
            </div>

            <div className='mt-6 flex items-center justify-between rounded-xl border border-border/60 bg-muted/30 p-3.5 font-mono text-xs'>
              <div className='flex items-center gap-2'>
                <Activity className='size-4 text-emerald-500' />
                <span>TTFT (First Token Latency)</span>
              </div>
              <span className='font-bold text-emerald-500'>&lt; 95ms (Direct Stream)</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
