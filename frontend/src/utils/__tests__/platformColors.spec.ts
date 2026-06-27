import { describe, expect, it } from 'vitest'
import {
  platformBadgeClass,
  platformGradientClass,
  platformTextClass
} from '../platformColors'

describe('platformColors', () => {
  it('Kiro 平台使用独立紫色主题，不复用 Anthropic 橙色', () => {
    expect(platformBadgeClass('kiro')).toContain('violet')
    expect(platformTextClass('kiro')).toContain('violet')
    expect(platformGradientClass('kiro')).toContain('from-violet-500')
    expect(platformGradientClass('kiro')).toContain('to-fuchsia-500')
    expect(platformBadgeClass('kiro')).not.toContain('orange')
  })

  it('Grok 平台使用独立 slate 主题', () => {
    expect(platformBadgeClass('grok')).toContain('slate')
    expect(platformTextClass('grok')).toContain('slate')
    expect(platformGradientClass('grok')).toContain('from-slate-500')
  })
})
