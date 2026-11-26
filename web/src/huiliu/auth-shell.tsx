/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/
import { Link } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'

type AuthShellProps = {
  children: React.ReactNode
}

/** Decorative lines converging into one, echoing the custom home page. */
function FlowLines() {
  return (
    <svg
      aria-hidden='true'
      viewBox='0 0 600 800'
      preserveAspectRatio='xMidYMid slice'
      className='pointer-events-none absolute inset-0 h-full w-full opacity-40'
    >
      {[120, 260, 400, 540, 680].map((y) => (
        <path
          key={y}
          d={`M-40 ${y} C 220 ${y}, 300 400, 640 400`}
          fill='none'
          stroke='currentColor'
          strokeWidth={y === 400 ? 1.6 : 1}
          strokeDasharray='6 14'
          strokeLinecap='round'
        />
      ))}
    </svg>
  )
}

function Brand(props: { className?: string }) {
  const { t } = useTranslation()
  const { systemName, logo, loading } = useSystemConfig()

  return (
    <Link
      to='/'
      className={`inline-flex items-center gap-2.5 transition-opacity hover:opacity-80 ${props.className ?? ''}`}
    >
      {loading ? (
        <Skeleton className='h-8 w-8 rounded-lg' />
      ) : (
        <img
          src={logo}
          alt={t('Logo')}
          className='h-8 w-8 rounded-lg object-cover'
        />
      )}
      {loading ? (
        <Skeleton className='h-5 w-24' />
      ) : (
        <span className='text-lg font-semibold tracking-tight'>
          {systemName}
        </span>
      )}
    </Link>
  )
}

/**
 * Two-column frame for the sign-in, sign-up, password and verification pages.
 * The forms themselves are unchanged upstream components rendered as children.
 */
export function AuthShell(props: AuthShellProps) {
  const { t } = useTranslation()
  const { systemName } = useSystemConfig()
  const year = new Date().getFullYear()

  return (
    <div className='bg-background grid min-h-svh lg:grid-cols-[minmax(0,5fr)_minmax(0,7fr)]'>
      <aside className='relative hidden overflow-hidden bg-[oklch(0.2_0.03_220)] text-[oklch(0.96_0.01_200)] lg:flex lg:flex-col lg:justify-between lg:p-12'>
        <div className='absolute inset-0 text-[oklch(0.78_0.13_182)]'>
          <FlowLines />
        </div>
        <Brand className='relative' />
        <div className='relative max-w-md space-y-4'>
          <p className='text-3xl leading-tight font-semibold tracking-tight xl:text-4xl'>
            {t('One key for leading AI models')}
          </p>
          <p className='text-base text-[oklch(0.82_0.02_210)]'>
            {t(
              'OpenAI-compatible API with usage-based billing. Every charge is listed in your logs.'
            )}
          </p>
        </div>
        <p className='relative text-sm text-[oklch(0.7_0.02_210)]'>
          © {year} {systemName}
        </p>
      </aside>

      <main className='flex min-h-svh flex-col'>
        <div className='flex items-center justify-between px-4 py-4 sm:px-8 sm:py-6'>
          <Brand className='lg:invisible' />
          <Link
            to='/'
            className='text-muted-foreground hover:text-foreground inline-flex items-center gap-1.5 text-sm transition-colors'
          >
            <ArrowLeft aria-hidden='true' className='h-4 w-4' />
            {t('Back to home')}
          </Link>
        </div>
        <div className='flex flex-1 items-center justify-center px-4 pb-12 sm:px-8'>
          <div className='w-full max-w-[420px] space-y-2'>{props.children}</div>
        </div>
      </main>
    </div>
  )
}
