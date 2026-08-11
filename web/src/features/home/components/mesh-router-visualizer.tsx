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
import { cn } from '@/lib/utils'

interface UpstreamNode {
  id: string
  name: string
  provider: string
  model: string
  latency: number
  costRatio: string
  status: 'healthy' | 'degraded' | 'failed'
  active: boolean
}

export function MeshRouterVisualizer({ className }: { className?: string }) {
  const { t } = useTranslation()

  const [selectedModel, setSelectedModel] = useState<string>('gpt-4o')
  const [isSimulatingFailure, setIsSimulatingFailure] = useState(false)
  const [routedNodeId, setRoutedNodeId] = useState('openai-main')
  const [rerouteLatency, setRerouteLatency] = useState<number | null>(null)
  const [pulseCount, setPulseCount] = useState(0)

  const initialNodes: UpstreamNode[] = [
    {
      id: 'openai-main',
      name: 'OpenAI Cluster-US',
      provider: 'OpenAI Direct',
      model: selectedModel,
      latency: 32,
      costRatio: '0.4x',
      status: isSimulatingFailure ? 'failed' : 'healthy',
      active: !isSimulatingFailure,
    },
    {
      id: 'azure-standby',
      name: 'Azure EastUS Standby',
      provider: 'Azure OpenAI',
      model: selectedModel,
      latency: 38,
      costRatio: '0.4x',
      status: 'healthy',
      active: isSimulatingFailure,
    },
    {
      id: 'deepseek-cluster',
      name: 'DeepSeek Direct High-QPS',
      provider: 'DeepSeek AI',
      model: 'deepseek-v3 / r1',
      latency: 16,
      costRatio: '0.2x',
      status: 'healthy',
      active: false,
    },
    {
      id: 'anthropic-mesh',
      name: 'Anthropic Direct Mesh',
      provider: 'Anthropic',
      model: 'claude-3-7-sonnet',
      latency: 44,
      costRatio: '0.4x',
      status: 'healthy',
      active: false,
    },
  ]

  const handleToggleFailure = () => {
    if (!isSimulatingFailure) {
      setIsSimulatingFailure(true)
      setRerouteLatency(2.1)
      setRoutedNodeId('azure-standby')
    } else {
      setIsSimulatingFailure(false)
      setRerouteLatency(null)
      setRoutedNodeId('openai-main')
    }
  }

  // Periodic visual packet pulse
  useEffect(() => {
    const interval = setInterval(() => {
      setPulseCount((prev) => (prev + 1) % 100)
    }, 1800)
    return () => clearInterval(interval)
  }, [])

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

      {/* Header bar: Live Telemetry */}
      <div className='relative z-10 mb-6 flex flex-wrap items-center justify-between gap-3 border-b border-border/60 pb-4'>
        <div className='flex items-center gap-3'>
          <div className='flex size-8 items-center justify-center rounded-lg border border-primary/30 bg-primary/10 text-primary'>
            <Activity className='size-4' />
          </div>
          <div>
            <div className='flex items-center gap-2'>
              <span className='font-mono text-xs font-bold tracking-wider uppercase'>
                {t('ArcMux Mesh Router')}
              </span>
              <span className='inline-flex items-center gap-1 rounded-md border border-emerald-500/30 bg-emerald-500/10 px-1.5 py-0.5 font-mono text-[10px] font-medium text-emerald-600 dark:text-emerald-400'>
                <span className='size-1.5 animate-pulse rounded-full bg-emerald-500' />
                ACTIVE
              </span>
            </div>
            <p className='text-muted-foreground text-xs'>
              {t('Sub-millisecond failover & multi-upstream multiplexer')}
            </p>
          </div>
        </div>

        {/* Model switcher tabs */}
        <div className='flex items-center gap-1.5 rounded-lg border border-border/60 bg-muted/40 p-1'>
          {['gpt-4o', 'claude-3-7-sonnet', 'deepseek-r1'].map((m) => (
            <button
              key={m}
              type='button'
              onClick={() => setSelectedModel(m)}
              className={cn(
                'rounded-md px-2.5 py-1 font-mono text-xs font-medium transition-colors',
                selectedModel === m
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
        <div className='flex flex-col gap-3 rounded-xl border border-border/60 bg-background/50 p-4 lg:col-span-4'>
          <div className='flex items-center justify-between border-b border-border/40 pb-2'>
            <span className='text-muted-foreground font-mono text-[11px] font-semibold uppercase tracking-wider'>
              {t('Inbound Traffic')}
            </span>
            <span className='font-mono text-[10px] text-muted-foreground/80'>
              POST /v1/chat/completions
            </span>
          </div>

          <div className='space-y-2 font-mono text-xs'>
            <div className='flex items-center justify-between rounded-lg bg-muted/40 px-3 py-2'>
              <span className='text-muted-foreground'>{t('Target Model')}</span>
              <span className='font-bold text-primary'>{selectedModel}</span>
            </div>
            <div className='flex items-center justify-between rounded-lg bg-muted/40 px-3 py-2'>
              <span className='text-muted-foreground'>{t('Protocol')}</span>
              <span className='text-foreground'>OpenAI / Responses / SSE</span>
            </div>
            <div className='flex items-center justify-between rounded-lg bg-muted/40 px-3 py-2'>
              <span className='text-muted-foreground'>{t('Billing Ratio')}</span>
              <span className='font-bold text-emerald-500'>0.40x (-60%)</span>
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
          <div className='relative flex size-16 items-center justify-center rounded-2xl border border-primary/40 bg-gradient-to-b from-primary/20 to-primary/5 shadow-lg shadow-primary/10'>
            <Cpu className='size-8 text-primary' />
            <span
              key={pulseCount}
              className='absolute inset-0 animate-ping rounded-2xl border border-primary/50 opacity-40'
            />
          </div>
          <div className='text-center'>
            <div className='font-mono text-xs font-bold'>{t('ArcMux Core')}</div>
            <div className='text-muted-foreground font-mono text-[10px]'>
              Overhead &lt; 0.35ms
            </div>
          </div>

          {rerouteLatency && (
            <div className='animate-in fade-in slide-in-from-top-1 inline-flex items-center gap-1.5 rounded-full border border-amber-500/30 bg-amber-500/10 px-2.5 py-0.5 font-mono text-[11px] font-semibold text-amber-600 dark:text-amber-400'>
              <Zap className='size-3 fill-current' />
              {t('Bypassed in')} {rerouteLatency}ms
            </div>
          )}
        </div>

        {/* Right: Upstream Nodes Matrix */}
        <div className='space-y-2.5 lg:col-span-5'>
          <div className='flex items-center justify-between px-1'>
            <span className='text-muted-foreground font-mono text-[11px] font-semibold uppercase tracking-wider'>
              {t('Multiplexed Upstream Nodes')}
            </span>
            <span className='font-mono text-[10px] text-muted-foreground'>
              4 Available
            </span>
          </div>

          {initialNodes.map((node) => {
            const isRouted = node.id === routedNodeId
            const isFailed = node.status === 'failed'

            return (
              <div
                key={node.id}
                className={cn(
                  'flex items-center justify-between rounded-xl border p-3 transition-all duration-200',
                  isFailed
                    ? 'border-red-500/40 bg-red-500/5 opacity-70'
                    : isRouted
                      ? 'border-primary/60 bg-primary/10 shadow-sm'
                      : 'border-border/60 bg-background/40 hover:border-border'
                )}
              >
                <div className='flex items-center gap-2.5'>
                  <div
                    className={cn(
                      'flex size-7 items-center justify-center rounded-lg text-xs font-bold',
                      isFailed
                        ? 'bg-red-500/20 text-red-500'
                        : isRouted
                          ? 'bg-primary text-primary-foreground'
                          : 'bg-muted text-muted-foreground'
                    )}
                  >
                    <Server className='size-3.5' />
                  </div>
                  <div>
                    <div className='flex items-center gap-2 font-mono text-xs font-semibold'>
                      <span>{node.name}</span>
                      {isRouted && (
                        <span className='inline-flex items-center gap-1 rounded bg-primary/20 px-1 py-0.2 font-mono text-[9px] font-bold text-primary uppercase'>
                          ROUTED
                        </span>
                      )}
                    </div>
                    <div className='text-muted-foreground font-mono text-[10px]'>
                      {node.provider}
                    </div>
                  </div>
                </div>

                <div className='text-right font-mono text-xs'>
                  {isFailed ? (
                    <span className='font-bold text-red-500'>
                      503 CIRCUIT OPEN
                    </span>
                  ) : (
                    <div className='flex items-center gap-2'>
                      <span className='text-muted-foreground text-[11px]'>
                        {node.latency}ms
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
