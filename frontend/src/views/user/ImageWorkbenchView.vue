<template>
  <AppLayout>
    <div class="-m-4 flex min-h-[calc(100dvh-4rem)] w-auto flex-col gap-4 bg-gray-50 p-4 dark:bg-dark-950 md:-m-6 md:p-5 lg:-m-8 lg:p-6">
      <div class="flex flex-col gap-3 border-b border-gray-200 pb-4 dark:border-dark-700 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <div class="flex items-center gap-2 text-sm font-medium text-primary-600 dark:text-primary-400">
            <Icon name="sparkles" size="sm" />
            <span>{{ t('imageWorkbench.eyebrow') }}</span>
          </div>
          <h1 class="mt-2 text-2xl font-semibold tracking-normal text-gray-900 dark:text-white">
            {{ t('imageWorkbench.title') }}
          </h1>
          <p class="mt-1 max-w-3xl text-sm leading-6 text-gray-600 dark:text-dark-300">
            {{ t('imageWorkbench.description') }}
          </p>
        </div>
        <button class="btn btn-secondary h-11" type="button" :disabled="keysLoading" @click="loadKeys">
          <Icon name="refresh" size="md" :class="keysLoading ? 'animate-spin' : ''" />
          <span class="ml-2">{{ t('common.refresh') }}</span>
        </button>
      </div>

      <div
        v-if="!keysLoading && usableKeys.length === 0"
        class="rounded-lg border border-dashed border-gray-300 bg-white px-6 py-12 dark:border-dark-600 dark:bg-dark-800"
      >
        <EmptyState
          :title="t('imageWorkbench.noKeyTitle')"
          :description="t('imageWorkbench.noKeyDescription')"
          :action-text="t('imageWorkbench.goKeys')"
          action-to="/keys"
        >
          <template #icon>
            <Icon name="key" size="xl" class="text-gray-400 dark:text-dark-300" />
          </template>
        </EmptyState>
      </div>

      <div v-else class="grid flex-1 gap-4 xl:min-h-0 xl:grid-cols-[340px_minmax(0,1fr)_340px] 2xl:grid-cols-[380px_minmax(0,1fr)_360px]">
        <section class="flex min-h-0 flex-col rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800 xl:max-h-[calc(100dvh-9rem)]">
          <div class="border-b border-gray-200 px-5 py-4 dark:border-dark-700">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('imageWorkbench.parameters') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">{{ t('imageWorkbench.parametersHint') }}</p>
          </div>

          <div class="min-h-0 flex-1 space-y-5 overflow-y-auto px-5 py-5">
            <label class="block">
              <span class="input-label mb-1.5 block">{{ t('imageWorkbench.apiKey') }}</span>
              <Select
                v-model="selectedKeyId"
                :options="keyOptions"
                :placeholder="t('imageWorkbench.selectKey')"
                :disabled="keysLoading || usableKeys.length === 0"
                searchable="auto"
              />
            </label>

            <div class="grid grid-cols-2 gap-3">
              <label class="block">
                <span class="input-label mb-1.5 block">{{ t('imageWorkbench.size') }}</span>
                <Select v-model="form.size" :options="sizeOptions" />
              </label>
              <label class="block">
                <span class="input-label mb-1.5 block">{{ t('imageWorkbench.count') }}</span>
                <Select v-model="form.n" :options="countOptions" />
              </label>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <label class="block">
                <span class="input-label mb-1.5 block">{{ t('imageWorkbench.quality') }}</span>
                <Select v-model="form.quality" :options="qualityOptions" />
              </label>
              <label class="block">
                <span class="input-label mb-1.5 block">{{ t('imageWorkbench.outputFormat') }}</span>
                <Select v-model="form.output_format" :options="formatOptions" />
              </label>
            </div>

            <div>
              <div class="mb-2 flex items-center justify-between gap-3">
                <span class="input-label">{{ t('imageWorkbench.referenceImages') }}</span>
                <button
                  v-if="referenceFiles.length"
                  type="button"
                  class="text-xs font-medium text-gray-500 hover:text-red-600 dark:text-dark-300 dark:hover:text-red-400"
                  @click="clearReferenceFiles"
                >
                  {{ t('imageWorkbench.clearReferences') }}
                </button>
              </div>
              <label
                class="flex min-h-[116px] cursor-pointer flex-col items-center justify-center rounded-lg border border-dashed border-gray-300 bg-gray-50 px-4 py-5 text-center transition hover:border-primary-400 hover:bg-primary-50/50 dark:border-dark-600 dark:bg-dark-900 dark:hover:border-primary-500/70 dark:hover:bg-primary-900/10"
              >
                <input
                  class="sr-only"
                  type="file"
                  accept="image/png,image/jpeg,image/webp"
                  multiple
                  @change="onReferenceInput"
                />
                <Icon name="upload" size="lg" class="text-gray-400 dark:text-dark-300" />
                <span class="mt-2 text-sm font-medium text-gray-700 dark:text-dark-100">
                  {{ t('imageWorkbench.uploadReference') }}
                </span>
                <span class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                  {{ t('imageWorkbench.uploadHint') }}
                </span>
              </label>

              <div v-if="referenceFiles.length" class="mt-3 grid grid-cols-3 gap-2">
                <div
                  v-for="file in referenceFiles"
                  :key="file.id"
                  class="group relative aspect-square overflow-hidden rounded-lg border border-gray-200 bg-gray-100 dark:border-dark-700 dark:bg-dark-900"
                >
                  <img :src="file.previewUrl" :alt="file.file.name" class="h-full w-full object-cover" />
                  <button
                    type="button"
                    class="absolute right-1 top-1 flex h-8 w-8 items-center justify-center rounded-md bg-white/90 text-gray-700 opacity-0 shadow-sm transition group-hover:opacity-100 dark:bg-dark-800/90 dark:text-dark-100"
                    :aria-label="t('imageWorkbench.removeReference')"
                    @click="removeReferenceFile(file.id)"
                  >
                    <Icon name="x" size="sm" />
                  </button>
                </div>
              </div>
            </div>

            <TextArea
              v-model="form.prompt"
              :label="t('imageWorkbench.prompt')"
              :placeholder="t('imageWorkbench.promptPlaceholder')"
              :rows="8"
              required
            />

            <button
              class="btn btn-primary h-12 w-full"
              type="button"
              :disabled="!canGenerate"
              @click="submit"
            >
              <Icon name="sparkles" size="md" :class="generating ? 'animate-pulse' : ''" />
              <span class="ml-2">{{ generating ? t('imageWorkbench.generating') : generateButtonLabel }}</span>
            </button>
          </div>
        </section>

        <section class="flex min-h-[560px] flex-col rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800 xl:min-h-0 xl:max-h-[calc(100dvh-9rem)]">
          <div class="flex flex-col gap-3 border-b border-gray-200 px-5 py-4 dark:border-dark-700 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('imageWorkbench.results') }}</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">{{ resultSummary }}</p>
            </div>
            <div class="flex items-center gap-2 text-xs text-gray-500 dark:text-dark-300">
              <Icon name="grid" size="sm" />
              <span>{{ activeEndpointLabel }}</span>
            </div>
          </div>

          <div class="min-h-0 flex-1 overflow-y-auto p-4 sm:p-5">
            <div v-if="generating" class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
              <div
                v-for="index in skeletonCount"
                :key="index"
                class="aspect-square animate-pulse rounded-lg border border-gray-200 bg-gray-100 dark:border-dark-700 dark:bg-dark-700"
              />
            </div>

            <div v-else-if="results.length === 0" class="flex min-h-[430px] items-center justify-center rounded-lg border border-dashed border-gray-300 bg-gray-50 dark:border-dark-600 dark:bg-dark-900 xl:min-h-full">
              <EmptyState
                :title="t('imageWorkbench.emptyCanvasTitle')"
                :description="t('imageWorkbench.emptyCanvasDescription')"
                :action-icon="false"
              >
                <template #icon>
                  <Icon name="grid" size="xl" class="text-gray-400 dark:text-dark-300" />
                </template>
              </EmptyState>
            </div>

            <div v-else class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
              <article
                v-for="image in results"
                :key="image.id"
                class="group overflow-hidden rounded-lg border bg-gray-50 transition dark:bg-dark-900"
                :class="selectedImageId === image.id ? 'border-primary-500 ring-2 ring-primary-500/20' : 'border-gray-200 dark:border-dark-700'"
              >
                <button
                  type="button"
                  class="block aspect-square w-full bg-white text-left dark:bg-dark-950"
                  @click="selectImage(image.id)"
                >
                  <img :src="image.src" :alt="image.prompt" class="h-full w-full object-contain" loading="lazy" />
                </button>
                <div class="space-y-3 p-3">
                  <div class="flex items-center justify-between gap-3">
                    <button
                      type="button"
                      class="rounded-md px-2 py-1 text-xs font-medium"
                      :class="selectedImageId === image.id ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-200' : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-dark-200 dark:hover:bg-dark-600'"
                      @click="selectImage(image.id)"
                    >
                      {{ selectedImageId === image.id ? t('imageWorkbench.selected') : t('imageWorkbench.select') }}
                    </button>
                    <div class="flex items-center gap-1">
                      <button
                        type="button"
                        class="flex h-9 w-9 items-center justify-center rounded-md text-gray-500 transition hover:bg-gray-100 hover:text-gray-800 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white"
                        :title="t('imageWorkbench.copyPrompt')"
                        @click="copyPrompt(image.revisedPrompt || image.prompt)"
                      >
                        <Icon name="copy" size="sm" />
                      </button>
                      <a
                        class="flex h-9 w-9 items-center justify-center rounded-md text-gray-500 transition hover:bg-gray-100 hover:text-gray-800 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white"
                        :href="image.src"
                        :download="`${downloadBaseName(image)}.${image.outputFormat}`"
                        :title="t('imageWorkbench.download')"
                      >
                        <Icon name="download" size="sm" />
                      </a>
                    </div>
                  </div>
                  <p class="line-clamp-2 text-xs leading-5 text-gray-500 dark:text-dark-300">
                    {{ image.revisedPrompt || image.prompt }}
                  </p>
                </div>
              </article>
            </div>
          </div>
        </section>

        <aside class="flex min-h-0 flex-col rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800 xl:max-h-[calc(100dvh-9rem)]">
          <div class="border-b border-gray-200 px-5 py-4 dark:border-dark-700">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('imageWorkbench.iteration') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">{{ t('imageWorkbench.iterationHint') }}</p>
          </div>

          <div class="min-h-0 flex-1 space-y-5 overflow-y-auto px-5 py-5">
            <div v-if="selectedImage" class="space-y-4">
              <div class="aspect-square overflow-hidden rounded-lg border border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-900">
                <img :src="selectedImage.src" :alt="selectedImage.prompt" class="h-full w-full object-contain" />
              </div>

              <div>
                <div class="mb-2 text-sm font-medium text-gray-900 dark:text-white">{{ t('imageWorkbench.revisedPrompt') }}</div>
                <p class="max-h-40 overflow-auto rounded-lg border border-gray-200 bg-gray-50 p-3 text-sm leading-6 text-gray-600 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-200">
                  {{ selectedImage.revisedPrompt || t('imageWorkbench.noRevisedPrompt') }}
                </p>
              </div>

              <div class="grid grid-cols-2 gap-2">
                <button class="btn btn-secondary h-11" type="button" @click="copyPrompt(selectedImage.revisedPrompt || selectedImage.prompt)">
                  <Icon name="copy" size="sm" />
                  <span class="ml-2">{{ t('imageWorkbench.copyPrompt') }}</span>
                </button>
                <button class="btn btn-secondary h-11" type="button" :disabled="!selectedImage.revisedPrompt" @click="useRevisedPrompt">
                  <Icon name="edit" size="sm" />
                  <span class="ml-2">{{ t('imageWorkbench.usePrompt') }}</span>
                </button>
              </div>

              <button class="btn btn-primary h-11 w-full" type="button" @click="setSelectedAsReference()">
                <Icon name="upload" size="sm" />
                <span class="ml-2">{{ t('imageWorkbench.setReference') }}</span>
              </button>

              <div class="space-y-2">
                <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t('imageWorkbench.remixActions') }}</div>
                <button
                  v-for="action in remixActions"
                  :key="action.key"
                  class="w-full rounded-lg border border-gray-200 px-3 py-3 text-left text-sm text-gray-700 transition hover:border-primary-300 hover:bg-primary-50 dark:border-dark-700 dark:text-dark-100 dark:hover:border-primary-700 dark:hover:bg-primary-900/20"
                  type="button"
                  @click="applyRemix(action.instruction)"
                >
                  <span class="font-medium">{{ action.label }}</span>
                  <span class="mt-1 block text-xs leading-5 text-gray-500 dark:text-dark-300">{{ action.description }}</span>
                </button>
              </div>
            </div>

            <div v-else class="rounded-lg border border-dashed border-gray-300 bg-gray-50 px-4 py-8 text-center dark:border-dark-600 dark:bg-dark-900">
              <Icon name="lightbulb" size="lg" class="mx-auto text-gray-400 dark:text-dark-300" />
              <p class="mt-3 text-sm font-medium text-gray-800 dark:text-dark-100">{{ t('imageWorkbench.noSelectionTitle') }}</p>
              <p class="mt-1 text-sm leading-6 text-gray-500 dark:text-dark-300">{{ t('imageWorkbench.noSelectionDescription') }}</p>
            </div>

            <div v-if="history.length" class="border-t border-gray-200 pt-5 dark:border-dark-700">
              <div class="mb-3 flex items-center justify-between">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('imageWorkbench.recentHistory') }}</h3>
                <button class="text-xs font-medium text-gray-500 hover:text-red-600 dark:text-dark-300 dark:hover:text-red-400" type="button" @click="clearHistory">
                  {{ t('imageWorkbench.clearHistory') }}
                </button>
              </div>
              <div class="grid grid-cols-4 gap-2">
                <button
                  v-for="item in history"
                  :key="item.id"
                  type="button"
                  class="aspect-square overflow-hidden rounded-md border border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-900"
                  :title="item.prompt"
                  @click="restoreHistory(item)"
                >
                  <img :src="item.src" :alt="item.prompt" class="h-full w-full object-cover" loading="lazy" />
                </button>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import TextArea from '@/components/common/TextArea.vue'
import Icon from '@/components/icons/Icon.vue'
import { keysAPI } from '@/api/keys'
import { editImage, generateImage } from '@/api/imageWorkbench'
import { useAppStore } from '@/stores/app'
import type { ApiKey } from '@/types'
import {
  IMAGE_WORKBENCH_PRESET_SIZES,
  isUsableImageWorkbenchKey,
  normalizeImageWorkbenchSize,
  validateImageWorkbenchFiles,
  type GeneratedWorkbenchImage
} from '@/utils/imageWorkbench'

interface WorkbenchForm {
  n: number
  output_format: string
  prompt: string
  quality: string
  size: string
}

interface ReferenceFile {
  file: File
  id: string
  previewUrl: string
}

interface WorkbenchImage extends GeneratedWorkbenchImage {
  createdAt: string
  endpoint: 'generations' | 'edits'
  id: string
  model: string
  outputFormat: string
  prompt: string
  quality: string
  size: string
}

const STORAGE_PARAMS = 'image_workbench_params'
const STORAGE_SELECTED_KEY = 'image_workbench_selected_key_id'
const STORAGE_HISTORY = 'image_workbench_history'
const HISTORY_LIMIT = 12
const IMAGE_WORKBENCH_MODEL = 'gpt-image-2'
const DEFAULT_FORM: WorkbenchForm = {
  n: 1,
  output_format: 'png',
  prompt: '',
  quality: 'auto',
  size: '1024x1024'
}

const { t } = useI18n()
const appStore = useAppStore()

const allKeys = ref<ApiKey[]>([])
const keysLoading = ref(false)
const generating = ref(false)
const selectedKeyId = ref<number | null>(null)
const referenceFiles = ref<ReferenceFile[]>([])
const results = ref<WorkbenchImage[]>([])
const history = ref<WorkbenchImage[]>([])
const selectedImageId = ref<string | null>(null)

const form = reactive<WorkbenchForm>({ ...DEFAULT_FORM })

const usableKeys = computed(() => allKeys.value.filter(isUsableImageWorkbenchKey))
const selectedKey = computed(() => usableKeys.value.find((key) => key.id === selectedKeyId.value) || null)
const selectedImage = computed(() => results.value.find((image) => image.id === selectedImageId.value) || null)
const isEditMode = computed(() => referenceFiles.value.length > 0)
const skeletonCount = computed(() => Math.max(1, Number(form.n) || 1))
const canGenerate = computed(() => Boolean(selectedKey.value?.key && form.prompt.trim() && !generating.value))
const generateButtonLabel = computed(() => isEditMode.value ? t('imageWorkbench.editButton') : t('imageWorkbench.generateButton'))
const activeEndpointLabel = computed(() => isEditMode.value ? '/v1/images/edits' : '/v1/images/generations')
const resultSummary = computed(() => {
  if (generating.value) {
    return t('imageWorkbench.generatingSummary')
  }
  if (!results.value.length) {
    return t('imageWorkbench.resultSummaryEmpty')
  }
  return t('imageWorkbench.resultSummary', { count: results.value.length })
})

const keyOptions = computed<SelectOption[]>(() =>
  usableKeys.value.map((key) => ({
    value: key.id,
    label: `${key.name} · ${key.group?.name || t('imageWorkbench.openaiGroup')}`
  }))
)

const sizeOptions = computed<SelectOption[]>(() =>
  IMAGE_WORKBENCH_PRESET_SIZES.map((size) => ({
    value: size,
    label: size === 'auto' ? t('imageWorkbench.auto') : size
  }))
)

const countOptions: SelectOption[] = [1, 2, 3, 4].map((value) => ({ value, label: String(value) }))
const qualityOptions = computed<SelectOption[]>(() => [
  { value: 'auto', label: t('imageWorkbench.auto') },
  { value: 'low', label: t('imageWorkbench.qualityLow') },
  { value: 'medium', label: t('imageWorkbench.qualityMedium') },
  { value: 'high', label: t('imageWorkbench.qualityHigh') }
])
const formatOptions: SelectOption[] = ['png', 'jpeg', 'webp'].map((value) => ({ value, label: value.toUpperCase() }))
const remixActions = computed(() => [
  {
    key: 'variation',
    label: t('imageWorkbench.remixVariation'),
    description: t('imageWorkbench.remixVariationDescription'),
    instruction: t('imageWorkbench.remixVariationInstruction')
  },
  {
    key: 'product',
    label: t('imageWorkbench.remixProduct'),
    description: t('imageWorkbench.remixProductDescription'),
    instruction: t('imageWorkbench.remixProductInstruction')
  },
  {
    key: 'background',
    label: t('imageWorkbench.remixBackground'),
    description: t('imageWorkbench.remixBackgroundDescription'),
    instruction: t('imageWorkbench.remixBackgroundInstruction')
  }
])

function restoreParams() {
  try {
    const raw = localStorage.getItem(STORAGE_PARAMS)
    if (raw) {
      const saved = JSON.parse(raw) as Partial<WorkbenchForm>
      form.n = Math.min(Math.max(Number(saved.n) || DEFAULT_FORM.n, 1), 4)
      form.output_format = typeof saved.output_format === 'string' ? saved.output_format : DEFAULT_FORM.output_format
      form.prompt = typeof saved.prompt === 'string' ? saved.prompt : DEFAULT_FORM.prompt
      form.quality = typeof saved.quality === 'string' ? saved.quality : DEFAULT_FORM.quality
      form.size = normalizeImageWorkbenchSize(String(saved.size || DEFAULT_FORM.size))
    }

    const keyId = Number(localStorage.getItem(STORAGE_SELECTED_KEY))
    if (Number.isFinite(keyId) && keyId > 0) {
      selectedKeyId.value = keyId
    }

    const historyRaw = localStorage.getItem(STORAGE_HISTORY)
    if (historyRaw) {
      history.value = (JSON.parse(historyRaw) as WorkbenchImage[]).slice(0, HISTORY_LIMIT)
    }
  } catch {
    // Ignore broken localStorage data.
  }
}

function persistParams() {
  try {
    localStorage.setItem(STORAGE_PARAMS, JSON.stringify({ ...form }))
  } catch {
    // ignore
  }
}

function persistHistory() {
  try {
    localStorage.setItem(STORAGE_HISTORY, JSON.stringify(history.value.slice(0, HISTORY_LIMIT)))
  } catch {
    // ignore
  }
}

async function loadKeys() {
  keysLoading.value = true
  try {
    const response = await keysAPI.list(1, 100)
    allKeys.value = response.items || []
    if (!selectedKey.value) {
      selectedKeyId.value = usableKeys.value[0]?.id ?? null
    }
  } catch {
    appStore.showError(t('imageWorkbench.keysLoadFailed'))
  } finally {
    keysLoading.value = false
  }
}

function onReferenceInput(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || [])
  input.value = ''
  if (!files.length) {
    return
  }

  const validation = validateImageWorkbenchFiles(files, {
    unsupportedType: (name) => t('imageWorkbench.unsupportedType', { name }),
    tooLarge: (name, maxMB) => t('imageWorkbench.tooLarge', { name, maxMB })
  })
  if (!validation.ok) {
    appStore.showError(validation.error || t('imageWorkbench.fileValidationFailed'))
    return
  }

  referenceFiles.value = [
    ...referenceFiles.value,
    ...files.map((file) => ({
      file,
      id: createWorkbenchId(),
      previewUrl: URL.createObjectURL(file)
    }))
  ]
}

function removeReferenceFile(id: string) {
  const file = referenceFiles.value.find((item) => item.id === id)
  if (file) {
    URL.revokeObjectURL(file.previewUrl)
  }
  referenceFiles.value = referenceFiles.value.filter((item) => item.id !== id)
}

function clearReferenceFiles() {
  for (const file of referenceFiles.value) {
    URL.revokeObjectURL(file.previewUrl)
  }
  referenceFiles.value = []
}

function selectImage(id: string) {
  selectedImageId.value = id
}

function toWorkbenchImages(images: GeneratedWorkbenchImage[], endpoint: 'generations' | 'edits'): WorkbenchImage[] {
  const now = new Date().toISOString()
  return images.map((image, index) => ({
    ...image,
    createdAt: now,
    endpoint,
    id: `${Date.now()}-${index}-${Math.random().toString(36).slice(2, 8)}`,
    model: IMAGE_WORKBENCH_MODEL,
    outputFormat: form.output_format,
    prompt: form.prompt.trim(),
    quality: form.quality,
    size: form.size
  }))
}

function addToHistory(images: WorkbenchImage[]) {
  history.value = [...images, ...history.value].slice(0, HISTORY_LIMIT)
  persistHistory()
}

async function submit() {
  if (!selectedKey.value?.key) {
    appStore.showError(t('imageWorkbench.selectKeyRequired'))
    return
  }
  if (!form.prompt.trim()) {
    appStore.showError(t('imageWorkbench.promptRequired'))
    return
  }

  generating.value = true
  selectedImageId.value = null
  try {
    form.size = normalizeImageWorkbenchSize(form.size)
    form.n = Math.min(Math.max(Number(form.n) || 1, 1), 4)
    const endpoint = isEditMode.value ? 'edits' : 'generations'
    const request = {
      ...form,
      apiKey: selectedKey.value.key,
      image: referenceFiles.value.map((item) => item.file),
      n: 1,
      prompt: form.prompt.trim()
    }
    const requestCount = form.n
    const responses = await Promise.allSettled(
      Array.from({ length: requestCount }, () => (
        endpoint === 'edits' ? editImage(request) : generateImage(request)
      ))
    )
    const failed = responses.filter((result) => result.status === 'rejected')
    const images = toWorkbenchImages(
      responses.flatMap((result) => result.status === 'fulfilled' ? result.value.images : []),
      endpoint
    )

    if (!images.length) {
      const firstError = failed[0]
      appStore.showError(
        firstError?.status === 'rejected' && firstError.reason instanceof Error
          ? firstError.reason.message
          : t('imageWorkbench.noImagesReturned')
      )
      return
    }

    results.value = images
    selectedImageId.value = images[0]?.id || null
    addToHistory(images)
    if (failed.length) {
      appStore.showWarning(t('imageWorkbench.partialGenerateSuccess', { count: images.length, failed: failed.length }))
    } else {
      appStore.showSuccess(t('imageWorkbench.generateSuccess'))
    }
  } catch (error) {
    appStore.showError(error instanceof Error ? error.message : t('imageWorkbench.generateFailed'))
  } finally {
    generating.value = false
  }
}

async function copyPrompt(prompt: string) {
  try {
    await navigator.clipboard.writeText(prompt)
    appStore.showSuccess(t('imageWorkbench.promptCopied'))
  } catch {
    appStore.showError(t('common.copyFailed'))
  }
}

function useRevisedPrompt() {
  if (selectedImage.value?.revisedPrompt) {
    form.prompt = selectedImage.value.revisedPrompt
  }
}

async function dataUrlToFile(dataUrl: string, fileName: string): Promise<File> {
  const response = await fetch(dataUrl)
  const blob = await response.blob()
  return new File([blob], fileName, { type: blob.type || `image/${form.output_format}` })
}

async function setSelectedAsReference(instruction?: string) {
  if (!selectedImage.value) {
    return
  }

  try {
    clearReferenceFiles()
    const file = await dataUrlToFile(selectedImage.value.src, `${downloadBaseName(selectedImage.value)}.${selectedImage.value.outputFormat}`)
    referenceFiles.value = [
      {
        file,
        id: createWorkbenchId(),
        previewUrl: URL.createObjectURL(file)
      }
    ]
    if (instruction) {
      form.prompt = appendInstruction(form.prompt || selectedImage.value.revisedPrompt || selectedImage.value.prompt, instruction)
    }
    await nextTick()
    appStore.showSuccess(t('imageWorkbench.referenceSet'))
  } catch {
    appStore.showError(t('imageWorkbench.referenceSetFailed'))
  }
}

function appendInstruction(prompt: string, instruction: string): string {
  const base = prompt.trim()
  return base ? `${base}\n\n${instruction}` : instruction
}

function createWorkbenchId(): string {
  return globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

async function applyRemix(instruction: string) {
  await setSelectedAsReference(instruction)
}

function restoreHistory(item: WorkbenchImage) {
  results.value = [item]
  selectedImageId.value = item.id
  form.prompt = item.prompt
  form.size = item.size || DEFAULT_FORM.size
  form.quality = item.quality || DEFAULT_FORM.quality
  form.output_format = item.outputFormat || DEFAULT_FORM.output_format
}

function clearHistory() {
  history.value = []
  persistHistory()
}

function clearObjectUrls() {
  for (const file of referenceFiles.value) {
    URL.revokeObjectURL(file.previewUrl)
  }
}

function downloadBaseName(image: WorkbenchImage): string {
  return `sub2api-image-${image.createdAt.slice(0, 19).replace(/[:T]/g, '-')}`
}

watch(form, persistParams, { deep: true })
watch(selectedKeyId, (value) => {
  try {
    if (value) {
      localStorage.setItem(STORAGE_SELECTED_KEY, String(value))
    } else {
      localStorage.removeItem(STORAGE_SELECTED_KEY)
    }
  } catch {
    // ignore
  }
})

onMounted(() => {
  restoreParams()
  loadKeys()
})

onBeforeUnmount(() => {
  clearObjectUrls()
})
</script>
