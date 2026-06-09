import { beforeEach, describe, expect, it, vi } from 'vitest'
import { generateImage } from '@/api/imageWorkbench'

vi.mock('@/i18n', () => ({
  getLocale: () => 'zh'
}))

describe('imageWorkbench API', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('uses gpt-image-2 by default and omits background from generation payload', async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ data: [{ b64_json: 'abc' }] }), {
      headers: { 'Content-Type': 'application/json' },
      status: 200
    }))
    vi.stubGlobal('fetch', fetchMock)

    await generateImage({
      apiKey: 'sk-image',
      n: 1,
      output_format: 'png',
      prompt: 'draw a product',
      quality: 'auto',
      size: '1024x1024'
    })

    const [, init] = fetchMock.mock.calls[0]
    const payload = JSON.parse(String(init.body))
    expect(payload).toEqual({
      model: 'gpt-image-2',
      n: 1,
      output_format: 'png',
      prompt: 'draw a product',
      quality: 'auto',
      size: '1024x1024'
    })
    expect(payload).not.toHaveProperty('background')
  })
})
