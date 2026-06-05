import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../CreateAccountModal.vue')
const componentSource = readFileSync(componentPath, 'utf8')

describe('CreateAccountModal platform lock', () => {
  it('locks the create form platform to the route-provided prop', () => {
    expect(componentSource).toContain('platform: AccountPlatform')
    expect(componentSource).toContain("platform: 'anthropic'")
    expect(componentSource).toContain('platform: props.platform')
    expect(componentSource).toContain('form.platform = props.platform')
  })

  it('does not render the in-dialog platform selector', () => {
    expect(componentSource).not.toContain('data-tour="account-form-platform"')
    expect(componentSource).not.toContain("@click=\"form.platform = 'anthropic'\"")
    expect(componentSource).not.toContain("@click=\"form.platform = 'openai'\"")
    expect(componentSource).not.toContain("@click=\"form.platform = 'gemini'\"")
    expect(componentSource).not.toContain("@click=\"form.platform = 'antigravity'\"")
  })
})
