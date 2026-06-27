import { describe, expect, it, vi } from 'vitest'

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn()
  })
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    kiro: {
      generateAuthUrl: vi.fn(),
      generateIDCAuthUrl: vi.fn(),
      exchangeCode: vi.fn(),
      refreshToken: vi.fn(),
      importToken: vi.fn()
    }
  }
}))

import { useKiroOAuth } from '@/composables/useKiroOAuth'
import { adminAPI } from '@/api/admin'

describe('useKiroOAuth.buildCredentials', () => {
  it('includes external_idp refresh fields in built credentials', () => {
    const oauth = useKiroOAuth()
    const creds = oauth.buildCredentials({
      access_token: 'at',
      refresh_token: 'rt',
      auth_method: 'external_idp',
      client_id: 'client-id',
      token_endpoint: 'https://login.microsoftonline.com/test/oauth2/v2.0/token',
      issuer_url: 'https://login.microsoftonline.com/test/v2.0',
      scopes: 'api://app/.default offline_access'
    })

    expect(creds.token_endpoint).toBe('https://login.microsoftonline.com/test/oauth2/v2.0/token')
    expect(creds.issuer_url).toBe('https://login.microsoftonline.com/test/v2.0')
    expect(creds.scopes).toBe('api://app/.default offline_access')
    expect(creds.client_id).toBe('client-id')
  })
})

describe('useKiroOAuth.validateRefreshToken', () => {
  it('forwards external_idp refresh fields to the backend payload', async () => {
    vi.mocked(adminAPI.kiro.refreshToken).mockResolvedValueOnce({
      access_token: 'new-access'
    })

    const oauth = useKiroOAuth()
    await oauth.validateRefreshToken({
      refreshToken: 'refresh-token',
      authMethod: 'external_idp',
      clientId: 'client-id',
      tokenEndpoint: 'https://login.microsoftonline.com/test/oauth2/v2.0/token',
      issuerUrl: 'https://login.microsoftonline.com/test/v2.0',
      scopes: 'api://app/.default offline_access'
    })

    expect(adminAPI.kiro.refreshToken).toHaveBeenCalledWith({
      refresh_token: 'refresh-token',
      auth_method: 'external_idp',
      provider: undefined,
      client_id: 'client-id',
      client_secret: undefined,
      token_endpoint: 'https://login.microsoftonline.com/test/oauth2/v2.0/token',
      issuer_url: 'https://login.microsoftonline.com/test/v2.0',
      scopes: 'api://app/.default offline_access',
      start_url: undefined,
      region: undefined,
      profile_arn: undefined,
      proxy_id: undefined
    })
  })
})
