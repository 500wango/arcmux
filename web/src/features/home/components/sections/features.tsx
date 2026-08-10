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
  Zap,
  Shield,
  Globe,
  Code,
  Gauge,
  DollarSign,
  Users,
  HeartHandshake,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'

interface FeaturesProps {
  className?: string
}

export function Features(_props: FeaturesProps) {
  const { t } = useTranslation()

  const features = [
    {
      id: 'fast',
      icon: <Zap className='size-5' strokeWidth={1.5} />,
      title: t('High Performance'),
      desc: t('Low-latency routing with automatic load balancing across providers'),
      tags: ['OpenAI', 'Claude', 'Gemini', 'DeepSeek', 'Qwen', 'Llama'],
    },
    {
      id: 'secure',
      icon: <Shield className='size-5' strokeWidth={1.5} />,
      title: t('Enterprise Security'),
      desc: t('Granular access control, key isolation, and audit trails out of the box'),
    },
    {
      id: 'global',
      icon: <Globe className='size-5' strokeWidth={1.5} />,
      title: t('Multi-Region'),
      desc: t('Deploy region-aware routing for resilient worldwide access'),
      subItems: [
        t('Load Balancing'),
        t('Rate Limiting'),
        t('Cost Tracking'),
      ],
    },
    {
      id: 'developer',
      icon: <Code className='size-5' strokeWidth={1.5} />,
      title: t('API Compatible'),
      desc: t('Drop-in compatible with OpenAI, Claude, and Gemini SDK conventions'),
      badge: t('Multi-protocol Compatible'),
    },
  ]

  const additionalFeatures = [
    {
      icon: <Gauge className='size-5' strokeWidth={1.5} />,
      title: t('Blazing Speed'),
      desc: t('Handle thousands of requests per second with intelligent distribution'),
    },
    {
      icon: <DollarSign className='size-5' strokeWidth={1.5} />,
      title: t('Transparent Billing'),
      desc: t('Pay-as-you-go with real-time usage monitoring'),
    },
    {
      icon: <Users className='size-5' strokeWidth={1.5} />,
      title: t('Team Collaboration'),
      desc: t('Multi-user management with flexible permission allocation'),
    },
    {
      icon: <HeartHandshake className='size-5' strokeWidth={1.5} />,
      title: t('Open Source'),
      desc: t('Community driven, self-hosted, and extensible'),
    },
  ]

  return (
    <section className='relative z-10 px-6 py-24 md:py-32'>
      <div className='mx-auto max-w-6xl'>
        {/* Section header */}
        <AnimateInView className='mx-auto mb-16 max-w-2xl text-center'>
          <p className='text-muted-foreground mb-3 text-xs font-semibold tracking-[0.12em] uppercase'>
            {t('Why Choose Us')}
          </p>
          <h2 className='text-2xl leading-tight font-bold tracking-tight md:text-3xl'>
            {t('Engineered for reliability,')}{' '}
            <span className='bg-gradient-to-r from-primary to-amber-500 bg-clip-text text-transparent'>
              {t('built to scale')}
            </span>
          </h2>
        </AnimateInView>

        {/* Feature cards grid */}
        <div className='grid gap-6 md:grid-cols-2'>
          {features.map((f, i) => (
            <AnimateInView
              key={f.id}
              delay={i * 100}
              animation='fade-up'
              className='tech-glass-card group relative overflow-hidden rounded-2xl p-7 transition-all duration-300 hover:-translate-y-1 hover:border-primary/40 hover:shadow-xl hover:shadow-primary/5'
            >
              {/* Subtle top corner ambient glow */}
              <div className='pointer-events-none absolute -top-12 -right-12 size-36 rounded-full bg-primary/10 opacity-0 blur-2xl transition-opacity duration-300 group-hover:opacity-100' />

              <div className='text-primary mb-4 flex size-12 items-center justify-center rounded-xl border border-primary/25 bg-primary/10 shadow-xs transition-transform duration-300 group-hover:scale-110'>
                {f.icon}
              </div>
              <h3 className='mb-2 text-lg font-bold tracking-tight'>{f.title}</h3>
              <p className='text-muted-foreground mb-5 text-sm leading-relaxed'>
                {f.desc}
              </p>
              {f.tags && (
                <div className='flex flex-wrap gap-1.5'>
                  {f.tags.map((tag) => (
                    <span
                      key={tag}
                      className='border-border/50 bg-muted/60 text-foreground/80 rounded-lg border px-2.5 py-1 text-xs font-medium backdrop-blur-xs transition-colors group-hover:border-primary/30'
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              )}
              {f.subItems && (
                <div className='space-y-2'>
                  {f.subItems.map((item) => (
                    <div
                      key={item}
                      className='text-muted-foreground flex items-center gap-2.5 text-xs font-medium'
                    >
                      <div className='bg-primary size-1.5 rounded-full shadow-[0_0_6px_rgba(234,88,12,0.6)]' />
                      {item}
                    </div>
                  ))}
                </div>
              )}
              {f.badge && (
                <div className='inline-flex items-center gap-1.5 rounded-lg border border-primary/30 bg-primary/10 px-3 py-1 text-xs font-medium text-primary shadow-xs'>
                  <Code className='size-3.5' />
                  {f.badge}
                </div>
              )}
            </AnimateInView>
          ))}
        </div>

        {/* Additional features row */}
        <div className='mt-12 grid grid-cols-2 gap-5 md:grid-cols-4'>
          {additionalFeatures.map((f, i) => (
            <AnimateInView
              key={f.title}
              delay={i * 100}
              animation='fade-up'
              className='tech-glass-card group rounded-xl p-5 text-center transition-all duration-300 hover:-translate-y-0.5 hover:border-primary/30 hover:shadow-md'
            >
              <div className='text-muted-foreground group-hover:text-primary mx-auto mb-3 flex size-10 items-center justify-center rounded-lg bg-primary/5 transition-all duration-300 group-hover:scale-110 group-hover:bg-primary/15'>
                {f.icon}
              </div>
              <h3 className='mb-1.5 text-sm font-bold tracking-tight'>{f.title}</h3>
              <p className='text-muted-foreground text-xs leading-relaxed'>
                {f.desc}
              </p>
            </AnimateInView>
          ))}
        </div>
      </div>
    </section>
  )
}