/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/
import { describe, expect, it } from 'vitest'

import type { NavGroup } from '@/components/layout/types'

import { planConsoleLayout } from '../console-plan'

const accountGroups: NavGroup[] = [
  {
    id: 'general',
    title: 'General',
    items: [
      { title: 'Overview', url: '/dashboard' },
      { title: 'API Keys', url: '/keys' },
      { title: 'Usage Logs', url: '/usage-logs' },
    ],
  },
  {
    id: 'personal',
    title: 'Personal',
    items: [{ title: 'Wallet', url: '/wallet' }],
  },
]

const adminGroup: NavGroup = {
  id: 'admin',
  title: 'Admin',
  items: [
    { title: 'Channels', url: '/channels' },
    { title: 'Users', url: '/users' },
  ],
}

describe('planConsoleLayout', () => {
  it('uses the shell with account tabs on an account page', () => {
    const plan = planConsoleLayout('/keys', accountGroups, false)

    expect(plan.kind).toBe('shell')
    if (plan.kind !== 'shell') return
    expect(plan.tabs.map((tab) => tab.url)).toEqual([
      '/dashboard',
      '/keys',
      '/usage-logs',
      '/wallet',
    ])
    expect(plan.adminHome).toBeUndefined()
  })

  it('matches nested account paths such as dashboard sections', () => {
    expect(
      planConsoleLayout('/dashboard/overview', accountGroups, false).kind
    ).toBe('shell')
  })

  it('offers the first admin page as admin home for administrators', () => {
    const plan = planConsoleLayout(
      '/wallet',
      [...accountGroups, adminGroup],
      false
    )

    expect(plan).toMatchObject({ kind: 'shell', adminHome: '/channels' })
    if (plan.kind !== 'shell') return
    expect(plan.tabs.map((tab) => tab.url)).not.toContain('/channels')
  })

  it('keeps the upstream layout on admin pages and nested workspaces', () => {
    const groups = [...accountGroups, adminGroup]

    expect(planConsoleLayout('/users', groups, false).kind).toBe('upstream')
    expect(planConsoleLayout('/system-settings/site', groups, true).kind).toBe(
      'upstream'
    )
  })

  it('keeps the upstream layout for pages outside the account tabs', () => {
    expect(planConsoleLayout('/playground', accountGroups, false).kind).toBe(
      'upstream'
    )
  })
})
