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
import { useQuery } from '@tanstack/react-query'
import {
  Activity,
  Cpu,
  Database,
  GitBranch,
  Network,
  ShieldCheck,
  Zap,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { PublicLayout } from '@/components/layout'
import { RichContent } from '@/components/rich-content'
import { Skeleton } from '@/components/ui/skeleton'
import { isHttpUrl, isLikelyHtml } from '@/lib/content-format'

import { getAboutContent } from './api'

function EmptyAboutState() {
  const { t } = useTranslation()
  const currentYear = new Date().getFullYear()

  return (
    <div className='mx-auto max-w-5xl py-12'>
      {/* Header telemetry badge */}
      <div className='mb-8 text-center'>
        <div className='mb-3 inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/5 px-3.5 py-1 font-mono text-xs font-semibold text-primary'>
          <Activity className='size-3.5' />
          <span>{t('Architecture & Specifications')}</span>
        </div>
        <h1 className='text-3xl font-extrabold tracking-tight text-foreground md:text-5xl'>
          {t('About ArcMux')}
        </h1>
        <p className='text-muted-foreground mx-auto mt-4 max-w-2xl text-sm leading-relaxed md:text-base'>
          {t(
            'A production-grade Go AI API Gateway and Multiplexer aggregating 40+ premier upstream AI providers behind a unified, resilient interface.'
          )}
        </p>
      </div>

      {/* Engineering Specs Grid */}
      <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-3'>
        <div className='rounded-2xl border border-border/80 bg-card/50 p-6 shadow-xs backdrop-blur-sm'>
          <div className='flex size-10 items-center justify-center rounded-xl border border-primary/30 bg-primary/10 text-primary'>
            <Cpu className='size-5' />
          </div>
          <h3 className='mt-4 font-mono text-sm font-bold'>{t('Go 1.22+ Core Engine')}</h3>
          <p className='text-muted-foreground mt-2 text-xs leading-relaxed'>
            {t('High-throughput non-blocking Gin runtime delivering sub-millisecond routing overhead.')}
          </p>
        </div>

        <div className='rounded-2xl border border-border/80 bg-card/50 p-6 shadow-xs backdrop-blur-sm'>
          <div className='flex size-10 items-center justify-center rounded-xl border border-emerald-500/30 bg-emerald-500/10 text-emerald-500'>
            <ShieldCheck className='size-5' />
          </div>
          <h3 className='mt-4 font-mono text-sm font-bold'>{t('Intelligent Failover')}</h3>
          <p className='text-muted-foreground mt-2 text-xs leading-relaxed'>
            {t('Live health probing and automated circuit breaker bypassing failed nodes in <3ms.')}
          </p>
        </div>

        <div className='rounded-2xl border border-border/80 bg-card/50 p-6 shadow-xs backdrop-blur-sm'>
          <div className='flex size-10 items-center justify-center rounded-xl border border-amber-500/30 bg-amber-500/10 text-amber-500'>
            <Zap className='size-5' />
          </div>
          <h3 className='mt-4 font-mono text-sm font-bold'>{t('Unbuffered SSE Pipeline')}</h3>
          <p className='text-muted-foreground mt-2 text-xs leading-relaxed'>
            {t('Zero gateway buffering with persistent HTTP keep-alive connection pooling.')}
          </p>
        </div>

        <div className='rounded-2xl border border-border/80 bg-card/50 p-6 shadow-xs backdrop-blur-sm'>
          <div className='flex size-10 items-center justify-center rounded-xl border border-blue-500/30 bg-blue-500/10 text-blue-500'>
            <Database className='size-5' />
          </div>
          <h3 className='mt-4 font-mono text-sm font-bold'>{t('Multi-DB Architecture')}</h3>
          <p className='text-muted-foreground mt-2 text-xs leading-relaxed'>
            {t('Native simultaneous support for SQLite, PostgreSQL 15+, and MySQL 8.0.')}
          </p>
        </div>

        <div className='rounded-2xl border border-border/80 bg-card/50 p-6 shadow-xs backdrop-blur-sm'>
          <div className='flex size-10 items-center justify-center rounded-xl border border-purple-500/30 bg-purple-500/10 text-purple-500'>
            <Network className='size-5' />
          </div>
          <h3 className='mt-4 font-mono text-sm font-bold'>{t('Universal Protocols')}</h3>
          <p className='text-muted-foreground mt-2 text-xs leading-relaxed'>
            {t('Native compatibility with /v1/chat, /v1/messages, /v1beta, and /v1/responses.')}
          </p>
        </div>

        <div className='rounded-2xl border border-border/80 bg-card/50 p-6 shadow-xs backdrop-blur-sm'>
          <div className='flex size-10 items-center justify-center rounded-xl border border-teal-500/30 bg-teal-500/10 text-teal-500'>
            <GitBranch className='size-5' />
          </div>
          <h3 className='mt-4 font-mono text-sm font-bold'>{t('Open Source & Extensible')}</h3>
          <p className='text-muted-foreground mt-2 text-xs leading-relaxed'>
            {t('Fully self-hostable with customizable billing formulas, ratios, and relaykit modules.')}
          </p>
        </div>
      </div>

      {/* Attribution & Legal Footer */}
      <div className='mt-12 rounded-2xl border border-border/60 bg-muted/20 p-6 text-center font-mono text-xs text-muted-foreground backdrop-blur-sm'>
        <div className='flex flex-wrap items-center justify-center gap-4 text-foreground'>
          <a
            href='https://github.com/500wango/arcmux'
            target='_blank'
            rel='noopener noreferrer'
            className='inline-flex items-center gap-1.5 font-semibold text-primary hover:underline'
          >
            <svg
              className='size-4 fill-current'
              viewBox='0 0 24 24'
              aria-hidden='true'
            >
              <path
                fillRule='evenodd'
                clipRule='evenodd'
                d='M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z'
              />
            </svg>
            <span>GitHub Repository</span>
          </a>
          <span>·</span>
          <span>AGPL-3.0 License</span>
          <span>·</span>
          <span>© {currentYear} ArcMux Contributors</span>
        </div>
        <p className='mt-3 text-[11px] text-muted-foreground/70'>
          Based on upstream One API / New API architectural foundations.
        </p>
      </div>
    </div>
  )
}

export function About() {
  const { t } = useTranslation()
  const { data, isLoading } = useQuery({
    queryKey: ['about-content'],
    queryFn: getAboutContent,
  })

  const rawContent = data?.data?.trim() ?? ''
  const hasContent = rawContent.length > 0
  const isUrl = hasContent && isHttpUrl(rawContent)
  const contentIsHtml = hasContent && isLikelyHtml(rawContent)

  if (isLoading) {
    return (
      <PublicLayout>
        <div className='mx-auto flex max-w-4xl flex-col gap-4 py-12'>
          <Skeleton className='h-8 w-[45%]' />
          <Skeleton className='h-4 w-full' />
          <Skeleton className='h-4 w-[90%]' />
          <Skeleton className='h-4 w-[80%]' />
        </div>
      </PublicLayout>
    )
  }

  if (!hasContent) {
    return (
      <PublicLayout>
        <EmptyAboutState />
      </PublicLayout>
    )
  }

  if (isUrl) {
    return (
      <PublicLayout showMainContainer={false}>
        <iframe
          src={rawContent}
          className='h-[calc(100vh-3.5rem)] w-full border-0'
          title={t('About')}
          sandbox='allow-forms allow-popups allow-popups-to-escape-sandbox allow-scripts'
        />
      </PublicLayout>
    )
  }

  if (contentIsHtml) {
    return (
      <PublicLayout showMainContainer={false}>
        <RichContent
          mode='html'
          htmlVariant='isolated'
          content={rawContent}
          className='prose-neutral dark:prose-invert max-w-none'
        />
      </PublicLayout>
    )
  }

  return (
    <PublicLayout>
      <div className='mx-auto max-w-6xl px-4 py-8'>
        <RichContent
          mode='markdown'
          content={rawContent}
          className='prose-neutral dark:prose-invert max-w-none'
        />
      </div>
    </PublicLayout>
  )
}
