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
import { CheckCircle2, Globe2 } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

interface ProviderItem {
  name: string
  category: 'flagship' | 'reasoning' | 'open' | 'enterprise'
  popularModels: string[]
  protocols: string[]
  features: string[]
}

export function ProviderMatrix({ className }: { className?: string }) {
  const { t } = useTranslation()
  const [filter, setFilter] = useState<'all' | 'flagship' | 'reasoning' | 'open' | 'enterprise'>('all')

  const providers: ProviderItem[] = [
    {
      name: 'OpenAI',
      category: 'flagship',
      popularModels: ['gpt-4o', 'gpt-4o-mini', 'o1', 'o3-mini'],
      protocols: ['/v1/chat/completions', '/v1/responses'],
      features: ['Structured Outputs', 'Tools', 'Vision'],
    },
    {
      name: 'Anthropic Claude',
      category: 'flagship',
      popularModels: ['claude-3-7-sonnet', 'claude-3-5-haiku', 'claude-3-opus'],
      protocols: ['/v1/messages', '/v1/chat/completions'],
      features: ['Extended Thinking', 'Prompt Cache', 'Vision'],
    },
    {
      name: 'DeepSeek',
      category: 'reasoning',
      popularModels: ['deepseek-v3', 'deepseek-r1'],
      protocols: ['/v1/chat/completions', '/v1/responses'],
      features: ['Reasoning Stream', 'Prefix Cache', 'Ultra Low-Cost'],
    },
    {
      name: 'Google Gemini',
      category: 'flagship',
      popularModels: ['gemini-2.5-pro', 'gemini-2.5-flash', 'gemini-2.0-flash-exp'],
      protocols: ['/v1beta', '/v1/chat/completions'],
      features: ['2M Context', 'Multimodal', 'Thinking Mode'],
    },
    {
      name: 'Alibaba Qwen',
      category: 'open',
      popularModels: ['qwen-2.5-max', 'qwen-2.5-coder-32b', 'qwq-32b-preview'],
      protocols: ['/v1/chat/completions'],
      features: ['Coding Benchmark', 'Tool Calling', 'Math'],
    },
    {
      name: 'Meta Llama',
      category: 'open',
      popularModels: ['llama-3.3-70b-instruct', 'llama-3.1-405b'],
      protocols: ['/v1/chat/completions'],
      features: ['Open Weights', 'Fast Token Flow', 'Long Context'],
    },
    {
      name: 'xAI Grok',
      category: 'reasoning',
      popularModels: ['grok-2-1212', 'grok-2-vision', 'grok-beta'],
      protocols: ['/v1/chat/completions'],
      features: ['Real-time Knowledge', 'Vision Reasoning'],
    },
    {
      name: 'Azure & AWS Bedrock',
      category: 'enterprise',
      popularModels: ['azure-openai-eastus', 'bedrock-claude-3-5', 'titan'],
      protocols: ['Enterprise IAM', 'Private Link'],
      features: ['99.99% SLA', 'High Concurrency', 'Dedicated QPS'],
    },
  ]

  const filteredProviders =
    filter === 'all'
      ? providers
      : providers.filter((p) => p.category === filter)

  return (
    <section className={cn('relative z-10 px-6 py-20 md:py-28', className)}>
      <div className='mx-auto max-w-6xl'>
        {/* Section Header */}
        <div className='mb-12 flex flex-col items-center text-center'>
          <div className='mb-3 inline-flex items-center gap-1.5 rounded-full border border-primary/20 bg-primary/5 px-3 py-1 font-mono text-xs font-semibold text-primary'>
            <Globe2 className='size-3.5' />
            {t('Universal Model Registry')}
          </div>
          <h2 className='text-3xl font-extrabold tracking-tight md:text-4xl'>
            {t('Unified Upstream Ecosystem')}
          </h2>
          <p className='text-muted-foreground mt-3 max-w-2xl text-sm leading-relaxed md:text-base'>
            {t(
              'Zero vendor lock-in. Switch, multiplex, and load-balance across 40+ premier AI architectures through one single API key.'
            )}
          </p>

          {/* Category Filter Pills */}
          <div className='mt-6 flex flex-wrap items-center justify-center gap-1.5 rounded-xl border border-border/60 bg-muted/30 p-1.5'>
            {[
              { id: 'all', label: t('All Providers') },
              { id: 'flagship', label: t('Flagship Tier') },
              { id: 'reasoning', label: t('Reasoning & R1') },
              { id: 'open', label: t('Open Weights') },
              { id: 'enterprise', label: t('Enterprise Cloud') },
            ].map((tab) => (
              <button
                key={tab.id}
                type='button'
                onClick={() => setFilter(tab.id as typeof filter)}
                className={cn(
                  'rounded-lg px-3.5 py-1.5 text-xs font-medium transition-all duration-200',
                  filter === tab.id
                    ? 'bg-background text-foreground shadow-xs'
                    : 'text-muted-foreground hover:text-foreground'
                )}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>

        {/* Matrix Grid */}
        <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
          {filteredProviders.map((provider) => (
            <div
              key={provider.name}
              className='group relative flex flex-col justify-between rounded-2xl border border-border/70 bg-card/40 p-5 shadow-xs backdrop-blur-sm transition-all duration-300 hover:border-primary/50 hover:bg-card/70 hover:shadow-md'
            >
              <div>
                <div className='flex items-center justify-between'>
                  <h3 className='font-mono text-sm font-bold tracking-tight text-foreground'>
                    {provider.name}
                  </h3>
                  <span className='inline-flex size-2 rounded-full bg-emerald-500/80 shadow-xs' />
                </div>

                <div className='mt-3 flex flex-wrap gap-1.5'>
                  {provider.popularModels.map((m) => (
                    <span
                      key={m}
                      className='rounded-md border border-border/60 bg-muted/40 px-2 py-0.5 font-mono text-[10px] text-muted-foreground transition-colors group-hover:text-foreground'
                    >
                      {m}
                    </span>
                  ))}
                </div>
              </div>

              <div className='mt-5 space-y-2 border-t border-border/40 pt-3'>
                <div className='flex items-center justify-between text-[11px] font-mono text-muted-foreground'>
                  <span>{t('Protocols')}:</span>
                  <span className='text-foreground'>{provider.protocols[0]}</span>
                </div>
                <div className='flex flex-wrap gap-1'>
                  {provider.features.map((f) => (
                    <span
                      key={f}
                      className='inline-flex items-center gap-1 font-mono text-[10px] text-primary/80'
                    >
                      <CheckCircle2 className='size-2.5 text-primary' />
                      {f}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
