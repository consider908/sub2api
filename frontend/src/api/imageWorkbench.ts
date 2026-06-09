import { getLocale } from '@/i18n'
import {
  getImageWorkbenchError,
  normalizeImageWorkbenchResult,
  type ImageWorkbenchResult
} from '@/utils/imageWorkbench'

export interface ImageWorkbenchRequest {
  apiKey: string
  image?: File[]
  n: number
  output_format: string
  prompt: string
  quality: string
  size: string
}

const IMAGE_WORKBENCH_MODEL = 'gpt-image-2'

async function parseResponsePayload(response: Response): Promise<unknown> {
  const contentType = response.headers.get('content-type') || ''
  if (contentType.includes('application/json')) {
    return response.json()
  }

  const text = await response.text()
  if (!text) {
    return {}
  }

  try {
    return JSON.parse(text)
  } catch {
    return { message: text }
  }
}

async function requestImages(
  endpoint: '/v1/images/generations' | '/v1/images/edits',
  init: RequestInit,
  outputFormat: string
): Promise<ImageWorkbenchResult> {
  const response = await fetch(endpoint, init)
  const payload = await parseResponsePayload(response)

  if (!response.ok) {
    throw new Error(getImageWorkbenchError(payload) || response.statusText || 'Image request failed')
  }

  return normalizeImageWorkbenchResult(payload, outputFormat)
}

function buildHeaders(apiKey: string, contentType?: string): Record<string, string> {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${apiKey}`,
    'Accept-Language': getLocale()
  }

  if (contentType) {
    headers['Content-Type'] = contentType
  }

  return headers
}

function buildJsonPayload(request: ImageWorkbenchRequest): Record<string, unknown> {
  return {
    model: IMAGE_WORKBENCH_MODEL,
    n: request.n,
    output_format: request.output_format,
    prompt: request.prompt,
    quality: request.quality,
    size: request.size
  }
}

export async function generateImage(request: ImageWorkbenchRequest): Promise<ImageWorkbenchResult> {
  return requestImages(
    '/v1/images/generations',
    {
      method: 'POST',
      headers: buildHeaders(request.apiKey, 'application/json'),
      body: JSON.stringify(buildJsonPayload(request))
    },
    request.output_format
  )
}

export async function editImage(request: ImageWorkbenchRequest): Promise<ImageWorkbenchResult> {
  const formData = new FormData()
  const payload = buildJsonPayload(request)
  for (const [key, value] of Object.entries(payload)) {
    formData.append(key, String(value))
  }
  for (const image of request.image || []) {
    formData.append('image', image)
  }

  return requestImages(
    '/v1/images/edits',
    {
      method: 'POST',
      headers: buildHeaders(request.apiKey),
      body: formData
    },
    request.output_format
  )
}

export const imageWorkbenchAPI = {
  editImage,
  generateImage
}

export default imageWorkbenchAPI
