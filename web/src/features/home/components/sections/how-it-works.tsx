/*
Copyright (C) 2023-2026 QuantumNous

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

For commercial licensing, please contact support@quantumnous.com
*/
import { Settings, Zap, BarChart3, ArrowRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'

export function HowItWorks() {
  const { t } = useTranslation()

  const steps = [
    {
      num: '01',
      title: t('Configure'),
      desc: t('Add upstream channels, manage keys, and set access policies'),
      icon: <Settings className='size-5' strokeWidth={1.5} />,
    },
    {
      num: '02',
      title: t('Connect'),
      desc: t('Call a single endpoint compatible with OpenAI, Claude, and Gemini formats'),
      icon: <Zap className='size-5' strokeWidth={1.5} />,
    },
    {
      num: '03',
      title: t('Monitor'),
      desc: t('Observe usage, cost, and latency through real-time dashboards'),
      icon: <BarChart3 className='size-5' strokeWidth={1.5} />,
    },
  ]

  return (
    <section className='relative z-10 px-6 py-24 md:py-32'>
      <div className='mx-auto max-w-6xl'>
        <AnimateInView className='mb-16 text-center'>
          <p className='text-muted-foreground mb-3 text-xs font-semibold tracking-[0.12em] uppercase'>
            {t('Get Started in Minutes')}
          </p>
          <h2 className='text-2xl font-bold tracking-tight md:text-3xl'>
            {t('Three steps to launch')}
          </h2>
        </AnimateInView>

        <div className='grid gap-8 md:grid-cols-3'>
          {steps.map((step, i) => (
            <AnimateInView
              key={step.num}
              delay={i * 150}
              animation='fade-up'
              className='relative'
            >
              <div className='tech-glass-card group relative overflow-hidden rounded-2xl p-7 transition-all duration-300 hover:-translate-y-1 hover:border-primary/40 hover:shadow-xl'>
                <div className='mb-5 flex items-center justify-between'>
                  <div className='text-primary flex size-12 items-center justify-center rounded-xl border border-primary/25 bg-primary/10 shadow-xs transition-transform duration-300 group-hover:scale-110'>
                    {step.icon}
                  </div>
                  <span className='text-muted-foreground/20 group-hover:text-primary/30 font-mono text-3xl font-black transition-colors duration-300 tabular-nums'>
                    {step.num}
                  </span>
                </div>
                <h3 className='mb-2 text-base font-bold tracking-tight'>{step.title}</h3>
                <p className='text-muted-foreground text-sm leading-relaxed'>
                  {step.desc}
                </p>
              </div>
              {/* Connector arrow (desktop) */}
              {i < steps.length - 1 && (
                <div className='hidden md:absolute md:top-1/2 md:-right-5 md:z-20 md:-translate-y-1/2 md:block'>
                  <div className='bg-muted/80 flex size-8 items-center justify-center rounded-full border border-border/60 shadow-xs'>
                    <ArrowRight className='text-muted-foreground size-4' />
                  </div>
                </div>
              )}
            </AnimateInView>
          ))}
        </div>
      </div>
    </section>
  )
}