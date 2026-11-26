/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/
import { Link, useLocation } from '@tanstack/react-router'
import { Shield } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { LanguageSwitcher } from '@/components/language-switcher'
import { AuthenticatedLayout } from '@/components/layout'
import { resolveSidebarView } from '@/components/layout/lib/sidebar-view-registry'
import type { NavLink } from '@/components/layout/types'
import { NotificationPopover } from '@/components/notification-popover'
import { AnimatedOutlet } from '@/components/page-transition'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { SkipToMain } from '@/components/skip-to-main'
import { ThemeSwitch } from '@/components/theme-switch'
import { SidebarProvider } from '@/components/ui/sidebar'
import { LayoutProvider } from '@/context/layout-provider'
import { SearchProvider } from '@/context/search-provider'
import { useNotifications } from '@/hooks/use-notifications'
import { useSidebarView } from '@/hooks/use-sidebar-view'
import { useSystemConfig } from '@/hooks/use-system-config'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

import { isActivePath, planConsoleLayout } from './console-plan'

function ConsoleShell(props: { tabs: NavLink[]; adminHome?: string }) {
  const { t } = useTranslation()
  const pathname = useLocation({ select: (l) => l.pathname })
  const { systemName, logo } = useSystemConfig()
  const notifications = useNotifications()

  return (
    <LayoutProvider>
      <SearchProvider>
        <SidebarProvider defaultOpen={false} className='flex-col'>
          <SkipToMain />
          <header className='border-border/60 bg-background/90 sticky top-0 z-40 border-b backdrop-blur-xl'>
            <div className='mx-auto flex h-14 max-w-7xl items-center gap-3 px-4 md:px-6'>
              <Link
                to='/'
                className='flex shrink-0 items-center gap-2 transition-opacity hover:opacity-80'
              >
                <img
                  src={logo}
                  alt=''
                  className='size-7 rounded-md object-cover'
                />
                <span className='hidden text-sm font-semibold tracking-tight sm:inline'>
                  {systemName}
                </span>
              </Link>
              <nav
                aria-label={t('Console')}
                className='-mb-px flex min-w-0 flex-1 items-stretch gap-1 self-stretch overflow-x-auto'
              >
                {props.tabs.map((item) => {
                  const url = String(item.url)
                  const active = isActivePath(pathname, url)
                  const Icon = item.icon
                  return (
                    <Link
                      key={url}
                      to={url}
                      aria-current={active ? 'page' : undefined}
                      className={cn(
                        'inline-flex shrink-0 items-center gap-1.5 border-b-2 px-2.5 text-sm font-medium whitespace-nowrap transition-colors',
                        active
                          ? 'border-primary text-foreground'
                          : 'text-muted-foreground hover:text-foreground border-transparent'
                      )}
                    >
                      {Icon && <Icon aria-hidden='true' className='size-4' />}
                      {t(item.title)}
                    </Link>
                  )
                })}
                {props.adminHome && (
                  <Link
                    to={props.adminHome}
                    className='text-muted-foreground hover:text-foreground inline-flex shrink-0 items-center gap-1.5 border-b-2 border-transparent px-2.5 text-sm font-medium whitespace-nowrap transition-colors'
                  >
                    <Shield aria-hidden='true' className='size-4' />
                    {t('Admin console')}
                  </Link>
                )}
              </nav>
              <div className='flex shrink-0 items-center gap-1'>
                <NotificationPopover
                  open={notifications.popoverOpen}
                  onOpenChange={notifications.setPopoverOpen}
                  unreadCount={notifications.unreadCount}
                  activeTab={notifications.activeTab}
                  onTabChange={notifications.setActiveTab}
                  notice={notifications.notice}
                  announcements={notifications.announcements}
                  loading={notifications.loading}
                />
                <LanguageSwitcher />
                <ThemeSwitch />
                <ProfileDropdown />
              </div>
            </div>
          </header>
          <div
            id='content'
            className='@container/content mx-auto flex h-[calc(100svh-3.5rem)] min-h-0 w-full max-w-7xl flex-col overflow-hidden'
          >
            <AnimatedOutlet />
          </div>
        </SidebarProvider>
      </SearchProvider>
    </LayoutProvider>
  )
}

/**
 * Console frame for signed-in pages. Account pages (overview, keys, logs,
 * wallet, profile, security) use the fork's top-tab shell; admin pages and
 * nested workspaces keep the upstream sidebar layout unchanged.
 */
export function ConsoleLayout() {
  const signedIn = useAuthStore((s) => !!s.auth.user)
  const pathname = useLocation({ select: (l) => l.pathname })
  const { navGroups } = useSidebarView()

  const plan = planConsoleLayout(
    pathname,
    navGroups,
    resolveSidebarView(pathname) !== null
  )

  if (!signedIn) return null
  if (plan.kind === 'upstream') return <AuthenticatedLayout />
  return <ConsoleShell tabs={plan.tabs} adminHome={plan.adminHome} />
}
