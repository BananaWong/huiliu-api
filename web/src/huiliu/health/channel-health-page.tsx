/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/
import { useQuery } from '@tanstack/react-query'
import { RefreshCw } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { ErrorState } from '@/components/error-state'
import { SectionPageLayout } from '@/components/layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
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

import {
  getChannelHealth,
  type ChannelHealthRow,
  type HealthStats,
} from './api'

const CHANNEL_ENABLED = 1
const CHANNEL_MANUALLY_DISABLED = 2

function StatsCell(props: {
  stats: HealthStats
  locale: Intl.LocalesArgument
}) {
  const { t } = useTranslation()
  const { stats, locale } = props
  if (!stats.requests) {
    return <span className='text-muted-foreground'>{t('No requests')}</span>
  }
  return (
    <div className='space-y-0.5'>
      <div>
        {formatPercent(stats.success_rate)}
        <span className='text-muted-foreground'>
          {' · '}
          {t('{{count}} requests', { count: stats.requests })}
        </span>
      </div>
      {stats.successes > 0 && (
        <div className='text-muted-foreground text-xs'>
          {t('Median {{p50}}s, P90 {{p90}}s', {
            p50: formatNumber(stats.p50_seconds, locale),
            p90: formatNumber(stats.p90_seconds, locale),
          })}
        </div>
      )}
    </div>
  )
}

function StatusCell(props: {
  row: ChannelHealthRow
  locale: Intl.LocalesArgument
}) {
  const { t } = useTranslation()
  const { row, locale } = props
  if (row.status === CHANNEL_ENABLED) {
    return <Badge variant='secondary'>{t('Enabled')}</Badge>
  }
  if (row.status === CHANNEL_MANUALLY_DISABLED) {
    return <Badge variant='outline'>{t('Manually disabled')}</Badge>
  }
  return (
    <div className='space-y-1'>
      <Badge variant='destructive'>{t('Automatically disabled')}</Badge>
      {row.next_retry_at ? (
        <div className='text-muted-foreground text-xs'>
          {t('Back on {{when}}', {
            when: formatTimestampRelative(row.next_retry_at, 'seconds', locale),
          })}
        </div>
      ) : null}
    </div>
  )
}

export function ChannelHealthPage() {
  const { t, i18n } = useTranslation()
  const locale = toIntlLocale(i18n.resolvedLanguage || i18n.language)
  const query = useQuery({
    queryKey: ['huiliu', 'channel-health'],
    queryFn: getChannelHealth,
    refetchInterval: 60_000,
  })
  const data = query.data

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>
        <span className='inline-flex min-w-0 items-center gap-2'>
          <span className='truncate'>{t('Channel Health')}</span>
          {data && (
            <Badge variant={data.enabled ? 'secondary' : 'outline'}>
              {data.enabled
                ? t('Automatic disable and recovery on')
                : t('Automatic disable and recovery off')}
            </Badge>
          )}
        </span>
      </SectionPageLayout.Title>
      <SectionPageLayout.Actions>
        <Button
          variant='outline'
          size='sm'
          onClick={() => query.refetch()}
          disabled={query.isFetching}
        >
          <RefreshCw className='size-4' />
          {t('Refresh')}
        </Button>
      </SectionPageLayout.Actions>
      <SectionPageLayout.Content>
        <p className='text-muted-foreground mb-4 text-sm'>
          {t(
            'Outcomes of real requests, counted per request. Caller errors such as invalid parameters do not count. Only channels with automatic disabling turned on are managed; decisions use the last {{minutes}} minutes.',
            { minutes: data?.window_minutes ?? 30 }
          )}
        </p>
        {query.isLoading && <Skeleton className='h-48 w-full' />}
        {query.isError && (
          <ErrorState
            title={t('We could not load channel health.')}
            description={
              query.error instanceof Error ? query.error.message : undefined
            }
          />
        )}
        {!query.isLoading && !query.isError && (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Channel')}</TableHead>
                <TableHead>{t('Status')}</TableHead>
                <TableHead>{t('Last hour')}</TableHead>
                <TableHead>{t('Last 24 hours')}</TableHead>
                <TableHead>{t('Last success')}</TableHead>
                <TableHead>{t('Last error')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(data?.channels ?? []).map((row) => (
                <TableRow key={row.id}>
                  <TableCell className='align-top'>
                    <div className='font-medium'>{row.name}</div>
                    <div className='text-muted-foreground text-xs'>
                      #{row.id} · {row.group}
                      {row.managed ? ` · ${t('Managed')}` : ''}
                    </div>
                    {row.models.length > 0 && (
                      <div className='text-muted-foreground mt-1 space-y-0.5 text-xs'>
                        {row.models.map((m) => (
                          <div key={m.model}>
                            {m.model}: {formatPercent(m.day.success_rate)} ·{' '}
                            {t('{{count}} requests', { count: m.day.requests })}
                          </div>
                        ))}
                      </div>
                    )}
                  </TableCell>
                  <TableCell className='align-top'>
                    <StatusCell row={row} locale={locale} />
                    {row.status !== CHANNEL_ENABLED && row.status_reason && (
                      <div className='text-muted-foreground mt-1 max-w-56 text-xs break-words'>
                        {row.status_reason}
                      </div>
                    )}
                  </TableCell>
                  <TableCell className='align-top'>
                    <StatsCell stats={row.hour} locale={locale} />
                    {row.hour.consecutive_failures > 0 && (
                      <div className='text-destructive text-xs'>
                        {t('{{count}} failures in a row', {
                          count: row.hour.consecutive_failures,
                        })}
                      </div>
                    )}
                  </TableCell>
                  <TableCell className='align-top'>
                    <StatsCell stats={row.day} locale={locale} />
                  </TableCell>
                  <TableCell className='text-muted-foreground align-top text-xs'>
                    {row.day.last_success_at
                      ? formatTimestampRelative(
                          row.day.last_success_at,
                          'seconds',
                          locale
                        )
                      : '-'}
                  </TableCell>
                  <TableCell className='align-top text-xs'>
                    {row.day.last_failure_at ? (
                      <div className='max-w-72 space-y-0.5'>
                        <div className='text-muted-foreground'>
                          {formatTimestampRelative(
                            row.day.last_failure_at,
                            'seconds',
                            locale
                          )}
                          {row.day.last_status_code
                            ? ` · ${row.day.last_status_code}`
                            : ''}
                        </div>
                        <div className='break-words'>{row.day.last_error}</div>
                      </div>
                    ) : (
                      '-'
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
