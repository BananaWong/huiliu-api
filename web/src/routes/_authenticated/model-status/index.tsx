/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/
import { createFileRoute } from '@tanstack/react-router'

import { ModelStatusPage } from '@/huiliu/health/model-status-page'

export const Route = createFileRoute('/_authenticated/model-status/')({
  component: ModelStatusPage,
})
