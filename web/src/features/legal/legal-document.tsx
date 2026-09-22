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
import { FileWarning, ShieldCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { PublicLayout } from '@/components/layout'
import { Footer } from '@/components/layout/components/footer'
import { RichContent } from '@/components/rich-content'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { isHttpUrl, isLikelyHtml } from '@/lib/content-format'

import type { LegalDocumentResponse } from './types'

type LegalDocumentProps = {
  title: string
  subtitle?: string
  badge?: string
  queryKey: string
  fetchDocument: () => Promise<LegalDocumentResponse>
  emptyMessage?: string
  defaultContent?: string
}

export function LegalDocument({
  title,
  subtitle,
  badge,
  queryKey,
  fetchDocument,
  emptyMessage,
  defaultContent,
}: LegalDocumentProps) {
  const { t } = useTranslation()
  const { data, isLoading } = useQuery({
    queryKey: [queryKey],
    queryFn: fetchDocument,
    staleTime: 10 * 60 * 1000,
  })

  const rawConfiguredContent = data?.data?.trim() ?? ''
  const isCustomConfigured = rawConfiguredContent.length > 0
  const isUrl = isCustomConfigured && isHttpUrl(rawConfiguredContent)
  const effectiveContent = isCustomConfigured
    ? rawConfiguredContent
    : (defaultContent?.trim() ?? '')
  const hasContent = effectiveContent.length > 0
  const contentIsHtml = hasContent && isLikelyHtml(effectiveContent)

  if (isLoading) {
    return (
      <PublicLayout showMainContainer={false}>
        <div className='container mx-auto min-h-[70vh] max-w-4xl space-y-6 px-4 py-12 pt-24'>
          <Skeleton className='h-8 w-[45%]' />
          <Skeleton className='h-4 w-full' />
          <Skeleton className='h-4 w-[90%]' />
          <Skeleton className='h-4 w-[80%]' />
        </div>
        <Footer />
      </PublicLayout>
    )
  }

  if (!hasContent) {
    return (
      <PublicLayout showMainContainer={false}>
        <div className='container mx-auto min-h-[70vh] max-w-2xl px-4 py-12 pt-24'>
          <Card className='border-dashed'>
            <CardHeader className='flex flex-row items-center gap-4'>
              <div className='bg-muted rounded-lg p-2'>
                <FileWarning className='text-muted-foreground h-5 w-5' />
              </div>
              <div className='space-y-1'>
                <CardTitle className='text-lg font-semibold'>{title}</CardTitle>
                <p className='text-muted-foreground text-sm'>
                  {data?.message ||
                    emptyMessage ||
                    t('No document configured yet.')}
                </p>
              </div>
            </CardHeader>
          </Card>
        </div>
        <Footer />
      </PublicLayout>
    )
  }

  if (isUrl) {
    return (
      <PublicLayout showMainContainer={false}>
        <div className='container mx-auto min-h-[70vh] max-w-2xl px-4 py-12 pt-24'>
          <Card>
            <CardHeader>
              <CardTitle>{title}</CardTitle>
            </CardHeader>
            <CardContent className='space-y-4'>
              <p className='text-muted-foreground text-sm'>
                {t(
                  'The administrator configured an external link for this document.'
                )}
              </p>
              <Button
                render={
                  <a
                    href={rawConfiguredContent}
                    target='_blank'
                    rel='noopener noreferrer'
                  />
                }
              >
                {t('View document')}
              </Button>
            </CardContent>
          </Card>
        </div>
        <Footer />
      </PublicLayout>
    )
  }

  return (
    <PublicLayout showMainContainer={false}>
      {contentIsHtml ? (
        <div className='min-h-[70vh] pt-20'>
          <RichContent
            mode='html'
            htmlVariant='isolated'
            content={effectiveContent}
          />
        </div>
      ) : (
        <div className='container mx-auto min-h-[70vh] max-w-4xl space-y-8 px-4 py-12 pt-24'>
          <div className='border-border/60 space-y-2 border-b pb-6'>
            <div className='border-primary/20 bg-primary/5 text-primary inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 font-mono text-xs font-semibold'>
              <ShieldCheck className='size-3.5' />
              <span>{badge ?? t('Legal & Compliance')}</span>
            </div>
            <h1 className='text-foreground text-3xl font-semibold tracking-tight md:text-4xl'>
              {title}
            </h1>
            {subtitle && (
              <p className='text-muted-foreground text-sm leading-relaxed'>
                {subtitle}
              </p>
            )}
          </div>

          <RichContent
            mode='markdown'
            content={effectiveContent}
            className='prose-neutral dark:prose-invert max-w-none'
          />
        </div>
      )}
      <Footer />
    </PublicLayout>
  )
}
