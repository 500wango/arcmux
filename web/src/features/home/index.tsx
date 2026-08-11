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
import { Code2 } from 'lucide-react'
import { useCallback, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'

import { PublicLayout } from '@/components/layout'
import { Footer } from '@/components/layout/components/footer'
import { RichContent } from '@/components/rich-content'
import { useTheme } from '@/context/theme-provider'
import { isLikelyHtml } from '@/lib/content-format'
import { useAuthStore } from '@/stores/auth-store'

import {
  BentoGrid,
  CodeIntegrationTabs,
  CTA,
  Hero,
  ProviderMatrix,
} from './components'
import { useHomePageContent } from './hooks'

export function Home() {
  const { i18n, t } = useTranslation()
  const iframeRef = useRef<HTMLIFrameElement>(null)
  const { resolvedTheme } = useTheme()
  const { auth } = useAuthStore()
  const isAuthenticated = !!auth.user
  const { content, isLoaded, isUrl } = useHomePageContent()

  const syncIframePreferences = useCallback(() => {
    try {
      iframeRef.current?.contentWindow?.postMessage(
        { themeMode: resolvedTheme },
        '*'
      )
      iframeRef.current?.contentWindow?.postMessage(
        { lang: i18n.language },
        '*'
      )
    } catch {
      // Cross-origin frames may reject access while navigating.
    }
  }, [i18n.language, resolvedTheme])

  useEffect(() => {
    if (isUrl) {
      syncIframePreferences()
    }
  }, [isUrl, syncIframePreferences])

  if (!isLoaded) {
    return (
      <PublicLayout showMainContainer={false}>
        <main className='flex min-h-screen items-center justify-center'>
          <div className='text-muted-foreground'>{t('Loading...')}</div>
        </main>
      </PublicLayout>
    )
  }

  if (content) {
    if (isUrl) {
      return (
        <PublicLayout showMainContainer={false}>
          <iframe
            ref={iframeRef}
            src={content}
            className='h-screen w-full border-none'
            title={t('Custom Home Page')}
            sandbox='allow-forms allow-popups allow-popups-to-escape-sandbox allow-scripts allow-top-navigation-by-user-activation'
            onLoad={syncIframePreferences}
          />
        </PublicLayout>
      )
    }

    const contentIsHtml = isLikelyHtml(content)

    if (contentIsHtml) {
      return (
        <PublicLayout showMainContainer={false}>
          <RichContent
            mode='html'
            htmlVariant='isolated'
            content={content}
            className='custom-home-content'
          />
        </PublicLayout>
      )
    }

    return (
      <PublicLayout>
        <div className='mx-auto max-w-6xl px-4 py-8'>
          <RichContent
            mode='markdown'
            content={content}
            className='custom-home-content'
          />
        </div>
      </PublicLayout>
    )
  }

  return (
    <PublicLayout showMainContainer={false}>
      {/* 1. Hero Section with Interactive Mesh Router Simulation */}
      <Hero isAuthenticated={isAuthenticated} />

      {/* 2. Supported Models & Upstream Provider Matrix */}
      <ProviderMatrix />

      {/* 3. Hardened Engine Architecture (Bento Grid) */}
      <BentoGrid />

      {/* 4. Multi-Language Developer Code Integration Section */}
      <section className='relative z-10 px-6 py-20 md:py-28'>
        <div className='mx-auto max-w-6xl'>
          <div className='mb-12 text-center'>
            <div className='mb-3 inline-flex items-center gap-1.5 rounded-full border border-primary/20 bg-primary/5 px-3 py-1 font-mono text-xs font-semibold text-primary'>
              <Code2 className='size-3.5' />
              <span>{t('Developer First')}</span>
            </div>
            <h2 className='text-3xl font-extrabold tracking-tight md:text-4xl'>
              {t('Drop-in Compatible with Every SDK')}
            </h2>
            <p className='text-muted-foreground mx-auto mt-3 max-w-2xl text-sm leading-relaxed md:text-base'>
              {t(
                'Only change the Base URL and API key. Works directly with official OpenAI, Anthropic, LangChain, and LlamaIndex packages.'
              )}
            </p>
          </div>

          <div className='mx-auto max-w-4xl'>
            <CodeIntegrationTabs />
          </div>
        </div>
      </section>

      {/* 5. Minimalist Call To Action */}
      <CTA isAuthenticated={isAuthenticated} />

      {/* 6. Footer */}
      <Footer />
    </PublicLayout>
  )
}
