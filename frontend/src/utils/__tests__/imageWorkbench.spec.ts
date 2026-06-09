import { describe, expect, it } from 'vitest'
import type { ApiKey } from '@/types'
import {
  extractGeneratedWorkbenchImages,
  getImageWorkbenchError,
  isUsableImageWorkbenchKey,
  normalizeImageWorkbenchSize,
  validateImageWorkbenchFiles
} from '@/utils/imageWorkbench'

function makeKey(overrides: Partial<ApiKey> = {}): ApiKey {
  return {
    id: 1,
    user_id: 1,
    key: 'sk-test',
    name: 'Test key',
    group_id: 1,
    status: 'active',
    ip_whitelist: [],
    ip_blacklist: [],
    last_used_at: null,
    quota: 0,
    quota_used: 0,
    expires_at: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    rate_limit_5h: 0,
    rate_limit_1d: 0,
    rate_limit_7d: 0,
    usage_5h: 0,
    usage_1d: 0,
    usage_7d: 0,
    window_5h_start: null,
    window_1d_start: null,
    window_7d_start: null,
    reset_5h_at: null,
    reset_1d_at: null,
    reset_7d_at: null,
    group: {
      id: 1,
      name: 'OpenAI Image',
      description: null,
      platform: 'openai',
      rate_multiplier: 1,
      is_exclusive: false,
      status: 'active',
      subscription_type: 'standard',
      daily_limit_usd: null,
      weekly_limit_usd: null,
      monthly_limit_usd: null,
      allow_image_generation: true,
      image_rate_independent: false,
      image_rate_multiplier: 1,
      image_price_1k: null,
      image_price_2k: null,
      image_price_4k: null,
      claude_code_only: false,
      fallback_group_id: null,
      fallback_group_id_on_invalid_request: null,
      require_oauth_only: false,
      require_privacy_set: false,
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z'
    },
    ...overrides
  }
}

const fileMessages = {
  unsupportedType: (name: string) => `unsupported:${name}`,
  tooLarge: (name: string, maxMB: number) => `large:${name}:${maxMB}`
}

describe('imageWorkbench utils', () => {
  it('keeps only active OpenAI keys with image generation enabled', () => {
    expect(isUsableImageWorkbenchKey(makeKey())).toBe(true)
    expect(isUsableImageWorkbenchKey(makeKey({ status: 'inactive' }))).toBe(false)
    expect(isUsableImageWorkbenchKey(makeKey({ group: { ...makeKey().group!, platform: 'anthropic' } }))).toBe(false)
    expect(isUsableImageWorkbenchKey(makeKey({ group: { ...makeKey().group!, allow_image_generation: false } }))).toBe(false)
    expect(isUsableImageWorkbenchKey(makeKey({ group: undefined }))).toBe(false)
  })

  it('extracts images from common response shapes', () => {
    expect(extractGeneratedWorkbenchImages({ data: [{ b64_json: 'abc', revised_prompt: 'better' }] }, 'png')).toEqual([
      { src: 'data:image/png;base64,abc', revisedPrompt: 'better' }
    ])
    expect(extractGeneratedWorkbenchImages({ data: [{ url: 'https://example.com/a.png' }] }, 'png')).toEqual([
      { src: 'https://example.com/a.png', revisedPrompt: undefined }
    ])
    expect(extractGeneratedWorkbenchImages({ images: ['raw-image'] }, 'webp')).toEqual([
      { src: 'data:image/webp;base64,raw-image' }
    ])
    expect(extractGeneratedWorkbenchImages({ output: [{ content: [{ image: 'nested' }] }] }, 'jpeg')).toEqual([
      { src: 'data:image/jpeg;base64,nested', revisedPrompt: undefined }
    ])
  })

  it('extracts errors from OpenAI and gateway-like payloads', () => {
    expect(getImageWorkbenchError({ error: { message: 'OpenAI error' } })).toBe('OpenAI error')
    expect(getImageWorkbenchError({ error: 'string error' })).toBe('string error')
    expect(getImageWorkbenchError({ message: 'message error' })).toBe('message error')
    expect(getImageWorkbenchError({ detail: 'detail error' })).toBe('detail error')
    expect(getImageWorkbenchError('nope')).toBeUndefined()
  })

  it('normalizes supported sizes and falls back for invalid values', () => {
    expect(normalizeImageWorkbenchSize('1024x1024')).toBe('1024x1024')
    expect(normalizeImageWorkbenchSize(' 2048 x 1152 ')).toBe('2048x1152')
    expect(normalizeImageWorkbenchSize('512×768')).toBe('512x768')
    expect(normalizeImageWorkbenchSize('32x32')).toBe('1024x1024')
    expect(normalizeImageWorkbenchSize('not-a-size', 'auto')).toBe('auto')
  })

  it('validates file type and size', () => {
    const ok = new File(['a'], 'ok.png', { type: 'image/png' })
    const unsupported = new File(['a'], 'bad.gif', { type: 'image/gif' })
    const tooLarge = new File([new Uint8Array(10 * 1024 * 1024 + 1)], 'large.png', { type: 'image/png' })

    expect(validateImageWorkbenchFiles([ok], fileMessages)).toEqual({ ok: true })
    expect(validateImageWorkbenchFiles([unsupported], fileMessages)).toEqual({ ok: false, error: 'unsupported:bad.gif' })
    expect(validateImageWorkbenchFiles([tooLarge], fileMessages)).toEqual({ ok: false, error: 'large:large.png:10' })
  })
})
