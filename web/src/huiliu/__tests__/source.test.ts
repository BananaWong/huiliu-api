/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/
import { describe, expect, it } from 'vitest'

import { SOURCE_REPOSITORY, sourceUrlForVersion } from '../source'

describe('sourceUrlForVersion', () => {
  it('links to the commit when the version ends with a commit hash', () => {
    expect(sourceUrlForVersion('huiliu-20261126-abc1234')).toBe(
      `${SOURCE_REPOSITORY}/tree/abc1234`
    )
  })

  it('links to the repository when the version has no commit hash', () => {
    expect(sourceUrlForVersion('v1.0.0-rc.51')).toBe(SOURCE_REPOSITORY)
    expect(sourceUrlForVersion(undefined)).toBe(SOURCE_REPOSITORY)
    expect(sourceUrlForVersion('')).toBe(SOURCE_REPOSITORY)
  })
})
