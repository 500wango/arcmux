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
import { type SVGProps, useId } from 'react'

import { cn } from '@/lib/utils'

type LogoVariant = 'icon' | 'mark' | 'full'

type LogoProps = SVGProps<SVGSVGElement> & {
  /** icon: app square; mark: wordmark; full: wordmark + tagline */
  variant?: LogoVariant
}

/**
 * ArcMux brand mark.
 * - icon: square app icon (default)
 * - mark: compact wordmark for headers
 * - full: wordmark + tagline for hero/footer
 */
export function Logo({
  className,
  variant = 'icon',
  ...props
}: LogoProps) {
  const reactId = useId().replace(/:/g, '')
  const gradientId = `arcmux-gradient-${reactId}`

  if (variant === 'icon') {
    return (
      <svg
        id='arcmux-logo-icon'
        viewBox='0 0 64 64'
        xmlns='http://www.w3.org/2000/svg'
        role='img'
        aria-label='ArcMux'
        className={cn('size-8', className)}
        {...props}
      >
        <title>ArcMux</title>
        <rect width='64' height='64' rx='14' fill='#020617' />
        <path
          d='M15 47 29 15h8l14 32h-9l-2.3-6H26.1l-2.3 6H15Z'
          fill='#f97316'
        />
        <path d='M28.5 34h8.8L33 22.8 28.5 34Z' fill='#fff7ed' />
        <path
          d='M14 30c12-6 24-10 38-12'
          fill='none'
          stroke='#facc15'
          strokeWidth='4'
          strokeLinecap='round'
        />
      </svg>
    )
  }

  const isFull = variant === 'full'
  return (
    <svg
      id={isFull ? 'arcmux-logo-full' : 'arcmux-logo-mark'}
      viewBox={isFull ? '0 0 300 130' : '0 0 300 90'}
      xmlns='http://www.w3.org/2000/svg'
      role='img'
      aria-label='ArcMux'
      className={cn(isFull ? 'h-auto w-48' : 'h-auto w-36', className)}
      {...props}
    >
      <title>ArcMux</title>
      <defs>
        <linearGradient
          id={gradientId}
          x1='0%'
          y1='100%'
          x2='100%'
          y2='0%'
        >
          <stop offset='0%' stopColor='#f97316' />
          <stop offset='100%' stopColor='#facc15' />
        </linearGradient>
      </defs>
      <g transform='skewX(-10) translate(20, 0)'>
        <text
          x='10'
          y='75'
          fontFamily="Inter, ui-sans-serif, system-ui, sans-serif"
          fontSize='58'
          fontWeight='800'
          fill={`url(#${gradientId})`}
          letterSpacing='-1'
        >
          Arc
        </text>
        <text
          x='105'
          y='75'
          fontFamily="Inter, ui-sans-serif, system-ui, sans-serif"
          fontSize='58'
          fontWeight='800'
          fill='currentColor'
          letterSpacing='-1'
        >
          Mux
        </text>
        <path
          d='M 18 55 Q 120 45 250 15'
          fill='none'
          stroke={`url(#${gradientId})`}
          strokeWidth='5'
          strokeLinecap='round'
        />
        <polygon points='250,15 238,12 245,22' fill='#facc15' />
      </g>
      {isFull ? (
        <text
          x='150'
          y='120'
          fontFamily="Inter, ui-sans-serif, system-ui, sans-serif"
          fontSize='11'
          fontWeight='600'
          fill='currentColor'
          opacity='0.55'
          letterSpacing='2'
          textAnchor='middle'
        >
          One Gateway for Every Model.
        </text>
      ) : null}
    </svg>
  )
}
