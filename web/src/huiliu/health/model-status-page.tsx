/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'

import { ErrorState } from '@/components/error-state'
import { SectionPageLayout } from '@/components/layout'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { toIntlLocale } from '@/i18n/languages'
import {
  formatNumber,
  formatPercent,
  formatTimestampRelative,
} from '@/lib/format'

import { getModelStatus, type ModelStatus } from './api'

const STATUS_BADGE: Record<
  ModelStatus,
  { labelKey: string; variant: 'secondary' | 'warning' | 'destructive' }
> = {
  normal: { labelKey: 'Running normally', variant: 'secondary' },
  degraded: { labelKey: 'Degraded', variant: 'warning' },
  unavailable: { labelKey: 'Out of service', variant: 'destructive' },
}

export function ModelStatusPage() {
  const { t, i18n } = useTranslation()
  const locale = toIntlLocale(i18n.resolvedLanguage || i18n.language)
  const query = useQuery({
    queryKey: ['huiliu', 'model-status'],
    queryFn: getModelStatus,
    refetchInterval: 60_000,
  })

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Model Status')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <p className='text-muted-foreground mb-4 text-sm'>
          {t(
            'Whether each model your account can use is working now, from real requests over the last 24 hours. Refreshes every minute.'
          )}
        </p>
        {query.isLoading && <Skeleton className='h-40 w-full' />}
        {query.isError && (
          <ErrorState
            title={t('We could not load model status.')}
            description={
              query.error instanceof Error ? query.error.message : undefined
            }
          />
        )}
        {!query.isLoading && !query.isError && (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Model')}</TableHead>
                <TableHead>{t('Status')}</TableHead>
                <TableHead>{t('Success rate (24 hours)')}</TableHead>
                <TableHead>{t('Median time')}</TableHead>
                <TableHead>{t('Last success')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(query.data ?? []).map((row) => {
                const badge = STATUS_BADGE[row.status] ?? STATUS_BADGE.normal
                return (
                  <TableRow key={row.model}>
                    <TableCell className='font-mono text-sm'>
                      {row.model}
                    </TableCell>
                    <TableCell>
                      <Badge variant={badge.variant}>{t(badge.labelKey)}</Badge>
                    </TableCell>
                    <TableCell>
                      {row.requests ? (
                        <>
                          {formatPercent(row.success_rate)}
                          <span className='text-muted-foreground'>
                            {' · '}
                            {t('{{count}} requests', { count: row.requests })}
                          </span>
                        </>
                      ) : (
                        <span className='text-muted-foreground'>
                          {t('No calls in the last 24 hours')}
                        </span>
                      )}
                    </TableCell>
                    <TableCell>
                      {row.requests && row.p50_seconds
                        ? `${formatNumber(row.p50_seconds, locale)}s`
                        : '-'}
                    </TableCell>
                    <TableCell className='text-muted-foreground'>
                      {row.last_success_at
                        ? formatTimestampRelative(
                            row.last_success_at,
                            'seconds',
                            locale
                          )
                        : '-'}
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        )}
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
