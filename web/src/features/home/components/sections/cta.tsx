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
import { Link } from '@tanstack/react-router'
import { ArrowRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'
import { Button } from '@/components/ui/button'

interface CTAProps {
  className?: string
  isAuthenticated?: boolean
}

export function CTA(props: CTAProps) {
  const { t } = useTranslation()

  if (props.isAuthenticated) {
    return null
  }

  return (
    <section className='relative z-10 overflow-hidden px-6 py-24 md:py-32'>
      {/* Ambient background aura */}
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 -z-10 opacity-20 dark:opacity-15'
        style={{
          background: [
            'radial-gradient(ellipse 60% 60% at 50% 50%, color-mix(in oklch, var(--primary) 40%, transparent) 0%, transparent 70%)',
            'radial-gradient(ellipse 40% 40% at 75% 60%, color-mix(in oklch, var(--chart-2) 30%, transparent) 0%, transparent 70%)',
          ].join(', '),
        }}
      />

      <div className='mx-auto max-w-6xl'>
        <AnimateInView
          className='tech-glass-card relative overflow-hidden rounded-3xl p-10 text-center shadow-2xl md:p-16'
          animation='scale-in'
        >
          {/* Decorative ambient radial glow inside card */}
          <div
            aria-hidden
            className='pointer-events-none absolute -top-24 -right-24 size-96 rounded-full opacity-30 blur-3xl'
            style={{
              background:
                'radial-gradient(circle, color-mix(in oklch, var(--primary) 70%, transparent) 0%, transparent 70%)',
            }}
          />

          <h2 className='relative text-3xl leading-tight font-extrabold tracking-tight md:text-5xl'>
            {t('Ready to unify')}
            <br />
            <span className='bg-gradient-to-r from-primary via-amber-500 to-yellow-500 bg-clip-text text-transparent drop-shadow-xs'>
              {t('your AI stack?')}
            </span>
          </h2>
          <p className='text-muted-foreground relative mx-auto mt-6 max-w-md text-base leading-relaxed'>
            {t('Spin up your own gateway and route every AI request through a single control plane.')}
          </p>
          <div className='relative mt-10 flex items-center justify-center gap-4'>
            <Button
              className='group h-11 rounded-xl px-6 font-semibold shadow-md transition-all duration-200 hover:shadow-lg hover:shadow-primary/25'
              render={<Link to='/sign-up' />}
            >
              {t('Get Started')}
              <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-1' />
            </Button>
            <Button
              variant='outline'
              className='h-11 rounded-xl border-border/60 bg-background/60 px-6 font-medium backdrop-blur-md transition-all duration-200 hover:border-border hover:bg-muted/60'
              render={<Link to='/pricing' />}
            >
              {t('View Pricing')}
            </Button>
          </div>
        </AnimateInView>
      </div>
    </section>
  )
}