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
import { Link } from '@tanstack/react-router'
import { Globe2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { usePricingData } from '@/features/pricing/hooks/use-pricing-data'
import { cn } from '@/lib/utils'

export function ProviderMatrix(props: { className?: string }) {
  const { t } = useTranslation()
  const { models, isLoading, error } = usePricingData()

  return (
    <section
      className={cn('relative z-10 px-6 py-20 md:py-28', props.className)}
    >
      <div className='mx-auto max-w-6xl'>
        <div className='mb-12 flex flex-col items-center text-center'>
          <Globe2 aria-hidden='true' className='text-primary mb-3 size-5' />
          <h2 className='text-3xl font-extrabold tracking-tight md:text-4xl'>
            {t('Available Models')}
          </h2>
          <p className='text-muted-foreground mt-3 max-w-2xl text-sm leading-relaxed md:text-base'>
            {t(
              'Models from this service’s pricing catalog. View the pricing page for the full list and rates.'
            )}
          </p>
          <Button
            className='mt-6'
            variant='outline'
            render={<Link to='/pricing' />}
          >
            {t('Model Square & Pricing')}
          </Button>
        </div>
        {isLoading && (
          <p role='status' className='text-muted-foreground text-center'>
            {t('Loading...')}
          </p>
        )}
        {error && (
          <p role='alert' className='text-muted-foreground text-center'>
            {t('Model catalog is unavailable. Please try again later.')}
          </p>
        )}
        {!isLoading && !error && models.length === 0 && (
          <p className='text-muted-foreground text-center'>
            {t('No models available')}
          </p>
        )}
        {!isLoading && !error && (
          <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
            {models.slice(0, 12).map((model) => (
              <article
                key={model.model_name}
                className='border-border/70 bg-card/40 min-w-0 rounded-2xl border p-5'
              >
                <h3 className='font-mono text-sm font-bold break-all'>
                  {model.model_name}
                </h3>
                {model.vendor_name && (
                  <p className='text-muted-foreground mt-2 text-sm'>
                    {model.vendor_name}
                  </p>
                )}
                {!!model.supported_endpoint_types?.length && (
                  <div className='border-border/40 mt-5 border-t pt-3'>
                    <p className='text-muted-foreground mb-2 text-xs'>
                      {t('Protocols')}
                    </p>
                    <div className='flex flex-wrap gap-2'>
                      {model.supported_endpoint_types.map((endpoint) => (
                        <code
                          key={endpoint}
                          className='text-muted-foreground text-xs break-all'
                        >
                          {endpoint}
                        </code>
                      ))}
                    </div>
                  </div>
                )}
              </article>
            ))}
          </div>
        )}
      </div>
    </section>
  )
}
