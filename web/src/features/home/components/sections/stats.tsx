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
import { useRef, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'

interface CounterProps {
  end: number
  suffix?: string
  prefix?: string
  duration?: number
  decimals?: number
}

function Counter(props: CounterProps) {
  const { end, suffix = '', prefix = '', duration = 1600, decimals = 0 } = props
  const ref = useRef<HTMLSpanElement>(null)
  const startedRef = useRef(false)

  const formatValue = useCallback(
    (v: number) =>
      decimals > 0 ? v.toFixed(decimals) : Math.round(v).toLocaleString(),
    [decimals]
  )

  const animate = useCallback(() => {
    const el = ref.current
    if (!el) return
    const start = performance.now()
    const step = (now: number) => {
      const progress = Math.min((now - start) / duration, 1)
      const eased = 1 - Math.pow(1 - progress, 3)
      el.textContent = `${prefix}${formatValue(eased * end)}${suffix}`
      if (progress < 1) requestAnimationFrame(step)
    }
    requestAnimationFrame(step)
  }, [end, duration, prefix, suffix, formatValue])

  useEffect(() => {
    const el = ref.current
    if (!el) return

    const mq = window.matchMedia('(prefers-reduced-motion: reduce)')
    if (mq.matches) {
      el.textContent = `${prefix}${formatValue(end)}${suffix}`
      return
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !startedRef.current) {
          startedRef.current = true
          animate()
          observer.unobserve(el)
        }
      },
      { threshold: 0.5 }
    )

    observer.observe(el)
    return () => observer.disconnect()
  }, [animate, end, prefix, suffix, formatValue])

  return (
    <span ref={ref} className='tabular-nums'>
      {prefix}0{suffix}
    </span>
  )
}

interface StatsProps {
  className?: string
}

interface StatItem {
  end: number
  suffix: string
  label: string
  decimals?: number
}

export function Stats(_props: StatsProps) {
  const { t } = useTranslation()

  const stats: StatItem[] = [
    { end: 50, suffix: '+', label: t('Upstream Providers') },
    { end: 100, suffix: '+', label: t('Billable Models') },
    { end: 50, suffix: '+', label: t('API Endpoints') },
    { end: 10, suffix: '+', label: t('Routing Strategies') },
  ]

  return (
    <div className='relative z-10 my-8'>
      <div className='mx-auto max-w-6xl px-6'>
        <div className='tech-glass-card divide-border/40 grid grid-cols-2 divide-x overflow-hidden rounded-2xl md:grid-cols-4'>
          {stats.map((s) => (
            <div
              key={s.label}
              className='group flex flex-col items-center py-8 text-center transition-colors duration-200 hover:bg-muted/30 md:py-10'
            >
              <span className='bg-gradient-to-r from-primary via-amber-500 to-yellow-500 bg-clip-text text-3xl font-extrabold tracking-tight text-transparent drop-shadow-xs md:text-4xl'>
                <Counter end={s.end} suffix={s.suffix} decimals={s.decimals} />
              </span>
              <span className='text-muted-foreground mt-2 text-xs font-medium tracking-wide'>
                {s.label}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

