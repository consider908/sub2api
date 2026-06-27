import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { PROVIDERS } from '@/constants/channelMonitor'

const root = resolve(__dirname, '../../../..')

function readSource(path: string): string {
  return readFileSync(resolve(root, path), 'utf8')
}

describe('channel monitor Kiro provider support', () => {
  it('includes kiro in shared provider constants', () => {
    expect(PROVIDERS).toContain('kiro')
  })

  it('shows kiro in monitor provider pickers and filters', () => {
    const formSource = readSource('components/admin/monitor/MonitorFormDialog.vue')
    const filtersSource = readSource('components/admin/monitor/MonitorFiltersBar.vue')
    const templateSource = readSource('components/admin/monitor/MonitorTemplateManagerDialog.vue')

    expect(formSource).toContain('PROVIDER_KIRO')
    expect(formSource).toContain("t('monitorCommon.providers.kiro')")
    expect(filtersSource).toContain('PROVIDER_KIRO')
    expect(filtersSource).toContain("t('monitorCommon.providers.kiro')")
    expect(templateSource).toContain('PROVIDER_KIRO')
    expect(templateSource).toContain('kiro: 0')
  })
})
