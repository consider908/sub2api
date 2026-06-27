import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })
})

describe('AppSidebar account platform navigation', () => {
  it('renders account management as platform child links', () => {
    expect(componentSource).toContain("import PlatformIcon from '@/components/common/PlatformIcon.vue'")
    expect(componentSource).toContain("const platformNavIcon = (platform: AccountPlatform)")
    expect(componentSource).toContain("path: '/admin/accounts',")
    expect(componentSource).toContain("expandOnly: true,")
    expect(componentSource).toContain("{ path: '/admin/accounts/anthropic', label: 'Anthropic', icon: platformNavIcon('anthropic') }")
    expect(componentSource).toContain("{ path: '/admin/accounts/openai', label: 'OpenAI', icon: platformNavIcon('openai') }")
    expect(componentSource).toContain("{ path: '/admin/accounts/gemini', label: 'Gemini', icon: platformNavIcon('gemini') }")
    expect(componentSource).toContain("{ path: '/admin/accounts/antigravity', label: 'Antigravity', icon: platformNavIcon('antigravity') }")
    expect(componentSource).toContain("{ path: '/admin/accounts/kiro', label: 'Kiro', icon: platformNavIcon('kiro') }")
    expect(componentSource).toContain("{ path: '/admin/accounts/grok', label: 'Grok', icon: platformNavIcon('grok') }")
  })
})
