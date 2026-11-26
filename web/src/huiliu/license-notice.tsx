/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/
import { useTranslation } from 'react-i18next'

import { useStatus } from '@/hooks/use-status'
import { cn } from '@/lib/utils'

import {
  SOURCE_REPOSITORY,
  UPSTREAM_REPOSITORY,
  sourceUrlForVersion,
} from './source'

type LicenseNoticeProps = {
  className?: string
}

/**
 * Legal notice required for this modified version: the upstream attribution
 * sentence and link (AGPL-3.0 section 7(b) terms in NOTICE), the license, and
 * where to get the source of the running build (section 13). Rendered by the
 * About page itself so editing the About option cannot remove it.
 */
export function LicenseNotice(props: LicenseNoticeProps) {
  const { t } = useTranslation()
  const { status } = useStatus()
  const linkClass = 'text-foreground underline-offset-4 hover:underline'

  return (
    <section
      aria-labelledby='license-notice-title'
      className={cn(
        'border-border/60 text-muted-foreground mx-auto w-full max-w-3xl border-t px-4 py-8 text-sm',
        props.className
      )}
    >
      <h2
        id='license-notice-title'
        className='text-foreground mb-3 text-sm font-semibold'
      >
        {t('Open source license')}
      </h2>
      <p className='mb-2'>
        Frontend design and development by New API contributors.
      </p>
      <ul className='space-y-1.5'>
        <li>
          {t('Upstream project')}:{' '}
          <a
            className={linkClass}
            href={UPSTREAM_REPOSITORY}
            target='_blank'
            rel='noopener noreferrer'
          >
            QuantumNous/new-api
          </a>
        </li>
        <li>
          {t('Source code of this site')}:{' '}
          <a
            className={linkClass}
            href={sourceUrlForVersion(status?.version)}
            target='_blank'
            rel='noopener noreferrer'
          >
            {SOURCE_REPOSITORY.replace('https://github.com/', '')}
          </a>{' '}
          (
          <a
            className={linkClass}
            href={`${SOURCE_REPOSITORY}/blob/huiliu/main/HUILIU_CHANGES.md`}
            target='_blank'
            rel='noopener noreferrer'
          >
            {t('Changes from upstream')}
          </a>
          )
        </li>
        <li>
          {t('License')}:{' '}
          <a
            className={linkClass}
            href={`${SOURCE_REPOSITORY}/blob/huiliu/main/LICENSE`}
            target='_blank'
            rel='noopener noreferrer'
          >
            GNU AGPL v3.0
          </a>
        </li>
      </ul>
    </section>
  )
}
