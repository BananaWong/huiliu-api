/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/

/** Public repository holding the source of this modified version (AGPL-3.0 section 13). */
export const SOURCE_REPOSITORY = 'https://github.com/BananaWong/huiliu-api'
export const UPSTREAM_REPOSITORY = 'https://github.com/QuantumNous/new-api'

/**
 * Link to the exact source of the running build. CI writes versions like
 * `huiliu-20261126-abc1234`; the trailing commit hash pins the tree.
 */
export function sourceUrlForVersion(
  version: string | undefined | null
): string {
  const commit = version?.match(/-([0-9a-f]{7,40})$/)?.[1]
  if (!commit) return SOURCE_REPOSITORY
  return `${SOURCE_REPOSITORY}/tree/${commit}`
}
