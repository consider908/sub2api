import type { ApiKey } from '@/types'

export const IMAGE_WORKBENCH_MAX_FILE_SIZE = 10 * 1024 * 1024
export const IMAGE_WORKBENCH_ACCEPTED_TYPES = new Set([
  'image/jpeg',
  'image/jpg',
  'image/png',
  'image/webp',
])

const MIN_CUSTOM_DIMENSION = 64
const MAX_CUSTOM_DIMENSION = 8192

export const IMAGE_WORKBENCH_PRESET_SIZES = [
  'auto',
  '1024x1024',
  '1536x1024',
  '1024x1536',
  '2048x2048',
  '2048x1152',
  '3840x2160',
  '2160x3840',
] as const

type UnknownRecord = Record<string, unknown>

export interface GeneratedWorkbenchImage {
  revisedPrompt?: string
  src: string
}

export interface ImageWorkbenchResult {
  background?: unknown
  created?: unknown
  endpoint?: string
  images: GeneratedWorkbenchImage[]
  model?: string
  outputFormat: string
  quality?: string
  size?: string
  usage?: unknown
}

export interface ImageWorkbenchValidationResult {
  ok: boolean
  error?: string
}

export function isUsableImageWorkbenchKey(key: ApiKey): boolean {
  return (
    key.status === 'active' &&
    key.group?.platform === 'openai' &&
    key.group?.allow_image_generation === true
  )
}

export function normalizeImageWorkbenchSize(value: string, fallback = '1024x1024'): string {
  const raw = String(value || '').trim()
  if ((IMAGE_WORKBENCH_PRESET_SIZES as readonly string[]).includes(raw)) {
    return raw
  }

  const normalized = raw.toLowerCase().replace(/\s+/g, '').replace(/x/g, 'x').replace(/×/g, 'x')
  const match = /^([1-9]\d{1,4})x([1-9]\d{1,4})$/.exec(normalized)
  if (!match) {
    return fallback
  }

  const width = Number(match[1])
  const height = Number(match[2])
  if (
    width < MIN_CUSTOM_DIMENSION ||
    width > MAX_CUSTOM_DIMENSION ||
    height < MIN_CUSTOM_DIMENSION ||
    height > MAX_CUSTOM_DIMENSION
  ) {
    return fallback
  }

  return `${width}x${height}`
}

export function validateImageWorkbenchFiles(
  files: File[],
  messages: {
    unsupportedType: (name: string) => string
    tooLarge: (name: string, maxMB: number) => string
  }
): ImageWorkbenchValidationResult {
  for (const file of files) {
    if (!IMAGE_WORKBENCH_ACCEPTED_TYPES.has(file.type)) {
      return { ok: false, error: messages.unsupportedType(file.name) }
    }
    if (file.size > IMAGE_WORKBENCH_MAX_FILE_SIZE) {
      return { ok: false, error: messages.tooLarge(file.name, IMAGE_WORKBENCH_MAX_FILE_SIZE / 1024 / 1024) }
    }
  }
  return { ok: true }
}

function isRecord(value: unknown): value is UnknownRecord {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function asString(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() ? value.trim() : undefined
}

function toImageSrc(value: unknown, outputFormat: string): string | undefined {
  const image = asString(value)
  if (!image) {
    return undefined
  }

  if (image.startsWith('data:image/') || image.startsWith('http://') || image.startsWith('https://')) {
    return image
  }

  return `data:image/${outputFormat};base64,${image}`
}

function collectImageFromRecord(record: UnknownRecord, outputFormat: string): GeneratedWorkbenchImage | undefined {
  const src =
    toImageSrc(record.b64_json, outputFormat) ||
    toImageSrc(record.url, outputFormat) ||
    toImageSrc(record.image, outputFormat) ||
    toImageSrc(record.base64, outputFormat) ||
    toImageSrc(record.result, outputFormat)

  if (!src) {
    return undefined
  }

  return {
    revisedPrompt: asString(record.revised_prompt) || asString(record.revisedPrompt),
    src,
  }
}

function collectFromArray(value: unknown, outputFormat: string): GeneratedWorkbenchImage[] {
  if (!Array.isArray(value)) {
    return []
  }

  return value.flatMap((item): GeneratedWorkbenchImage[] => {
    const src = toImageSrc(item, outputFormat)
    if (src) {
      return [{ src }]
    }
    if (!isRecord(item)) {
      return []
    }

    const image = collectImageFromRecord(item, outputFormat)
    if (image) {
      return [image]
    }

    return [
      ...collectFromArray(item.data, outputFormat),
      ...collectFromArray(item.images, outputFormat),
      ...collectFromArray(item.output, outputFormat),
      ...collectFromArray(item.content, outputFormat),
    ]
  })
}

export function extractGeneratedWorkbenchImages(payload: unknown, outputFormat: string): GeneratedWorkbenchImage[] {
  if (!isRecord(payload)) {
    return []
  }

  const image = collectImageFromRecord(payload, outputFormat)
  const images = [
    ...collectFromArray(payload.data, outputFormat),
    ...collectFromArray(payload.images, outputFormat),
    ...collectFromArray(payload.output, outputFormat),
    ...collectFromArray(payload.content, outputFormat),
  ]

  return image ? [image, ...images] : images
}

export function getImageWorkbenchError(payload: unknown): string | undefined {
  if (!isRecord(payload)) {
    return undefined
  }

  if (isRecord(payload.error)) {
    return asString(payload.error.message) || asString(payload.error.type)
  }

  return (
    asString(payload.error) ||
    asString(payload.message) ||
    asString(payload.msg) ||
    asString(payload.detail)
  )
}

export function normalizeImageWorkbenchResult(payload: unknown, outputFormat: string): ImageWorkbenchResult {
  const record = isRecord(payload) ? payload : {}
  return {
    background: record.background,
    created: record.created,
    endpoint: asString(record.endpoint),
    images: extractGeneratedWorkbenchImages(payload, outputFormat),
    model: asString(record.model),
    outputFormat,
    quality: asString(record.quality),
    size: asString(record.size),
    usage: record.usage,
  }
}
