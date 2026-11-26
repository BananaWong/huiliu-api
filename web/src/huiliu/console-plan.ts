/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/
import type { NavGroup, NavLink } from '@/components/layout/types'

const ADMIN_GROUP_ID = 'admin'

function linksOf(groups: NavGroup[]): NavLink[] {
  return groups
    .flatMap((group) => group.items)
    .filter((item): item is NavLink => typeof item.url === 'string')
}

export function isActivePath(pathname: string, url: string): boolean {
  return pathname === url || pathname.startsWith(`${url}/`)
}

export type ConsolePlan =
  | { kind: 'upstream' }
  | { kind: 'shell'; tabs: NavLink[]; adminHome?: string }

/**
 * Decide the frame for a signed-in page: account pages get the top-tab
 * shell; admin pages, nested workspaces and anything unknown keep the
 * upstream sidebar layout.
 */
export function planConsoleLayout(
  pathname: string,
  groups: NavGroup[],
  inNestedView: boolean
): ConsolePlan {
  const adminLinks = linksOf(groups.filter((g) => g.id === ADMIN_GROUP_ID))
  const accountLinks = linksOf(groups.filter((g) => g.id !== ADMIN_GROUP_ID))
  const onAdminPage =
    inNestedView ||
    adminLinks.some((item) => isActivePath(pathname, String(item.url)))
  const onAccountPage = accountLinks.some((item) =>
    isActivePath(pathname, String(item.url))
  )
  if (onAdminPage || !onAccountPage) return { kind: 'upstream' }
  return {
    kind: 'shell',
    tabs: accountLinks,
    adminHome: adminLinks[0] ? String(adminLinks[0].url) : undefined,
  }
}
