/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/
import { api } from '@/lib/api'
import { requireServerSuccess } from '@/lib/server-error-message'

/** Request outcomes over a window; times are unix seconds, durations seconds. */
export type HealthStats = {
  requests: number
  successes: number
  failures: number
  success_rate: number
  p50_seconds: number
  p90_seconds: number
  last_success_at: number
  last_failure_at: number
  last_error?: string
  last_status_code?: number
  consecutive_failures: number
}

export type ChannelHealthRow = {
  id: number
  name: string
  status: number
  managed: boolean
  group: string
  status_reason?: string
  status_time?: number
  next_retry_at?: number
  hour: HealthStats
  day: HealthStats
  models: { model: string; day: HealthStats }[]
}

export type ChannelHealth = {
  enabled: boolean
  window_minutes: number
  channels: ChannelHealthRow[]
}

export type ModelStatus = 'normal' | 'degraded' | 'unavailable'

export type ModelStatusRow = {
  model: string
  status: ModelStatus
  requests: number
  success_rate: number
  p50_seconds: number
  last_success_at: number
}

type Envelope<T> = { success: boolean; message?: string; data?: T }

export async function getChannelHealth(): Promise<ChannelHealth> {
  const res = await api.get<Envelope<ChannelHealth>>(
    '/api/huiliu/channel-health'
  )
  const data = requireServerSuccess(res.data).data
  return {
    enabled: data?.enabled ?? false,
    window_minutes: data?.window_minutes ?? 0,
    channels: data?.channels ?? [],
  }
}

export async function getModelStatus(): Promise<ModelStatusRow[]> {
  const res = await api.get<Envelope<ModelStatusRow[]>>(
    '/api/huiliu/model-status'
  )
  return requireServerSuccess(res.data).data ?? []
}
