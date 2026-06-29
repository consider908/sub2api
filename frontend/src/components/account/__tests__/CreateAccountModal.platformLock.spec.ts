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

  it('restores Kiro apikey creation alongside OAuth methods', () => {
    expect(componentSource).toContain("form.platform === 'kiro'")
    expect(componentSource).toContain("@click=\"accountCategory = 'apikey'\"")
    expect(componentSource).toContain("{{ t('admin.accounts.types.kiroApikey') }}")
    expect(componentSource).toContain("kiroAuthMethod = 'google'")
    expect(componentSource).toContain("kiroAuthMethod = 'github'")
    expect(componentSource).toContain("kiroAuthMethod = 'idc'")
    expect(componentSource).toContain("kiroAuthMethod = 'import'")
  })

  it('exposes and dispatches Kiro refresh-token creation', () => {
    expect(componentSource).toContain(":show-refresh-token-option=\"form.platform === 'openai' || form.platform === 'antigravity' || form.platform === 'kiro'\"")
    expect(componentSource).toContain("} else if (form.platform === 'kiro') {\n    handleKiroValidateRT(rt)")
    expect(componentSource).toContain('const handleKiroValidateRT = async')
    expect(componentSource).toContain('const credentials = buildKiroCredentials(tokenInfo)')
    expect(componentSource).toContain("platform: 'kiro'")
    expect(componentSource).toContain("type: 'oauth'")
  })

  it('collects IDC client credentials before Kiro refresh-token validation', () => {
    expect(componentSource).toContain('const kiroIDCClientId = ref')
    expect(componentSource).toContain('const kiroIDCClientSecret = ref')
    expect(componentSource).toContain('showKiroIDCRefreshTokenFields')
    expect(componentSource).toContain('v-model="kiroIDCClientId"')
    expect(componentSource).toContain('v-model="kiroIDCClientSecret"')
    expect(componentSource).toContain('clientId: isIDC ? idcClientId : undefined')
    expect(componentSource).toContain('clientSecret: isIDC ? idcClientSecret : undefined')
  })

  it('uses Grok OAuth title instead of Claude fallback', () => {
    expect(componentSource).toContain("if (form.platform === 'grok') return t('admin.accounts.oauth.grok.title')")
  })

  it('blocks Kiro IDC refresh-token validation when required fields are missing', () => {
    expect(componentSource).toContain("kiroOAuth.error.value = t('admin.accounts.oauth.kiro.pleaseEnterRefreshToken')")
    expect(componentSource).toContain("kiroOAuth.error.value = t('admin.accounts.oauth.kiro.pleaseEnterClientId')")
    expect(componentSource).toContain("kiroOAuth.error.value = t('admin.accounts.oauth.kiro.pleaseEnterClientSecret')")
    expect(componentSource).toContain("appStore.showError(kiroOAuth.error.value)")
    expect(componentSource).toContain('return')
  })
})
