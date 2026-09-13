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
  AlertTriangle,
  CheckCircle2,
  Cpu,
  RefreshCw,
  Server,
  Zap,
} from 'lucide-react'
import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { usePricingData } from '@/features/pricing/hooks/use-pricing-data'
import { cn } from '@/lib/utils'

export function MeshRouterVisualizer({ className }: { className?: string }) {
  const { t } = useTranslation()
  const { models, isLoading, error } = usePricingData()
  const exampleModels = models
    .filter((model) => model.supported_endpoint_types?.includes('openai'))
    .slice(0, 3)
  const [selectedModelName, setSelectedModel] = useState('')
  const selectedModel =
    exampleModels.find((model) => model.model_name === selectedModelName) ??
    exampleModels[0]
  const [isSimulatingFailure, setIsSimulatingFailure] = useState(false)
  const [pulseCount, setPulseCount] = useState(0)
  const routedNodeId = isSimulatingFailure ? 'backup' : 'primary'
  const initialNodes = [
    {
      id: 'primary',
      name: `${t('Channel')} 1`,
      status: isSimulatingFailure ? 'failed' : 'healthy',
    },
    { id: 'backup', name: `${t('Channel')} 2`, status: 'healthy' },
  ]

  const handleToggleFailure = () => setIsSimulatingFailure((failed) => !failed)

  // Periodic visual packet pulse
  useEffect(() => {
    const interval = setInterval(() => {
      setPulseCount((prev) => (prev + 1) % 100)
    }, 1800)
    return () => clearInterval(interval)
  }, [])

  if (isLoading) return <p role='status'>{t('Loading...')}</p>
  if (error) {
    return (
      <p role='alert'>
        {t('Model catalog is unavailable. Please try again later.')}
      </p>
    )
  }
  if (!selectedModel) {
    return (
      <p>{t('No model with Chat Completions support is currently listed.')}</p>
    )
  }

  return (
    <div
      className={cn(
        'group relative overflow-hidden rounded-2xl border border-border/80 bg-card/60 p-5 shadow-2xl backdrop-blur-xl transition-all duration-300 md:p-7',
        className
      )}
    >
      {/* Background subtle micro-grid */}
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 opacity-[0.03] dark:opacity-[0.05]'
        style={{
          backgroundImage:
            'radial-gradient(circle, currentColor 1px, transparent 1px)',
          backgroundSize: '16px 16px',
        }}
      />

      {/* Routing example, not service telemetry */}
      <div className='border-border/60 relative z-10 mb-6 flex flex-wrap items-center justify-between gap-3 border-b pb-4'>
        <div className='flex items-center gap-3'>
          <div className='border-primary/30 bg-primary/10 text-primary flex size-8 items-center justify-center rounded-lg border'>
            <Activity className='size-4' />
          </div>
          <div>
            <div className='flex items-center gap-2'>
              <span className='font-mono text-xs font-bold tracking-wider uppercase'>
                {t('ArcMux Mesh Router')}
              </span>
              <span className='inline-flex items-center gap-1 rounded-md border border-emerald-500/30 bg-emerald-500/10 px-1.5 py-0.5 font-mono text-[10px] font-medium text-emerald-600 dark:text-emerald-400'>
                <span className='size-1.5 animate-pulse rounded-full bg-emerald-500' />
                {t('Example')}
              </span>
            </div>
            <p className='text-muted-foreground text-xs'>
              {t(
                'Interactive routing example. No live traffic, health, or performance data is shown.'
              )}
            </p>
          </div>
        </div>

        {/* Model switcher tabs */}
        <div className='border-border/60 bg-muted/40 flex items-center gap-1.5 rounded-lg border p-1'>
          {exampleModels.map(({ model_name: m }) => (
            <button
              key={m}
              type='button'
              onClick={() => setSelectedModel(m)}
              className={cn(
                'rounded-md px-2.5 py-1 font-mono text-xs font-medium transition-colors',
                selectedModel.model_name === m
                  ? 'bg-background text-foreground shadow-xs'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              {m}
            </button>
          ))}
        </div>
      </div>

      {/* Interactive Visualizer Canvas */}
      <div className='relative z-10 grid gap-6 lg:grid-cols-12 lg:items-center'>
        {/* Left: Client Inbound */}
        <div className='border-border/60 bg-background/50 flex flex-col gap-3 rounded-xl border p-4 lg:col-span-4'>
          <div className='border-border/40 flex items-center justify-between border-b pb-2'>
            <span className='text-muted-foreground font-mono text-[11px] font-semibold tracking-wider uppercase'>
              {t('Inbound Traffic')}
            </span>
            <span className='text-muted-foreground/80 font-mono text-[10px]'>
              POST /v1/chat/completions
            </span>
          </div>

          <div className='space-y-2 font-mono text-xs'>
            <div className='bg-muted/40 flex items-center justify-between rounded-lg px-3 py-2'>
              <span className='text-muted-foreground'>{t('Target Model')}</span>
              <span className='text-primary font-bold'>
                {selectedModel.model_name}
              </span>
            </div>
            <div className='bg-muted/40 flex items-center justify-between rounded-lg px-3 py-2'>
              <span className='text-muted-foreground'>{t('Protocol')}</span>
              <span className='text-foreground'>/v1/chat/completions</span>
            </div>
          </div>

          <div className='pt-2'>
            <Button
              size='sm'
              variant={isSimulatingFailure ? 'destructive' : 'outline'}
              onClick={handleToggleFailure}
              className='w-full gap-2 text-xs font-medium shadow-xs transition-all'
            >
              {isSimulatingFailure ? (
                <>
                  <RefreshCw className='size-3.5 animate-spin' />
                  {t('Restore Primary Channel')}
                </>
              ) : (
                <>
                  <AlertTriangle className='size-3.5 text-amber-500' />
                  {t('Simulate Upstream 503 Outage')}
                </>
              )}
            </Button>
          </div>
        </div>

        {/* Center: ArcMux Gateway Engine Hub */}
        <div className='flex flex-col items-center justify-center gap-2 py-2 lg:col-span-3 lg:py-0'>
          <div className='border-primary/40 from-primary/20 to-primary/5 shadow-primary/10 relative flex size-16 items-center justify-center rounded-2xl border bg-gradient-to-b shadow-lg'>
            <Cpu className='text-primary size-8' />
            <span
              key={pulseCount}
              className='border-primary/50 absolute inset-0 animate-ping rounded-2xl border opacity-40'
            />
          </div>
          <div className='text-center'>
            <div className='font-mono text-xs font-bold'>
              {t('ArcMux Core')}
            </div>
            <div className='text-muted-foreground font-mono text-[10px]'>
              {t('Load Balancing')}
            </div>
          </div>

          {isSimulatingFailure && (
            <div className='animate-in fade-in slide-in-from-top-1 inline-flex items-center gap-1.5 rounded-full border border-amber-500/30 bg-amber-500/10 px-2.5 py-0.5 font-mono text-[11px] font-semibold text-amber-600 dark:text-amber-400'>
              <Zap className='size-3 fill-current' />
              {t('Retry')}
            </div>
          )}
        </div>

        {/* Right: Upstream Nodes Matrix */}
        <div className='space-y-2.5 lg:col-span-5'>
          <div className='flex items-center justify-between px-1'>
            <span className='text-muted-foreground font-mono text-[11px] font-semibold tracking-wider uppercase'>
              {t('Multiplexed Upstream Nodes')}
            </span>
            <span className='text-muted-foreground font-mono text-[10px]'>
              {t('Example')}
            </span>
          </div>

          {initialNodes.map((node) => {
            const isRouted = node.id === routedNodeId
            const isFailed = node.status === 'failed'
            let nodeClass =
              'border-border/60 bg-background/40 hover:border-border'
            let iconClass = 'bg-muted text-muted-foreground'
            if (isFailed) {
              nodeClass = 'border-red-500/40 bg-red-500/5 opacity-70'
              iconClass = 'bg-red-500/20 text-red-500'
            } else if (isRouted) {
              nodeClass = 'border-primary/60 bg-primary/10 shadow-sm'
              iconClass = 'bg-primary text-primary-foreground'
            }

            return (
              <div
                key={node.id}
                className={cn(
                  'flex items-center justify-between rounded-xl border p-3 transition-all duration-200',
                  nodeClass
                )}
              >
                <div className='flex items-center gap-2.5'>
                  <div
                    className={cn(
                      'flex size-7 items-center justify-center rounded-lg text-xs font-bold',
                      iconClass
                    )}
                  >
                    <Server className='size-3.5' />
                  </div>
                  <div>
                    <div className='flex items-center gap-2 font-mono text-xs font-semibold'>
                      <span>{node.name}</span>
                      {isRouted && (
                        <span className='bg-primary/20 py-0.2 text-primary inline-flex items-center gap-1 rounded px-1 font-mono text-[9px] font-bold uppercase'>
                          {t('Active')}
                        </span>
                      )}
                    </div>
                    <div className='text-muted-foreground font-mono text-[10px]'>
                      {t('Example')}
                    </div>
                  </div>
                </div>

                <div className='text-right font-mono text-xs'>
                  {isFailed ? (
                    <span className='font-bold text-red-500'>HTTP 503</span>
                  ) : (
                    <div className='flex items-center gap-2'>
                      <span className='text-muted-foreground text-[11px]'>
                        {t('Healthy')}
                      </span>
                      <CheckCircle2 className='size-3.5 text-emerald-500' />
                    </div>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
