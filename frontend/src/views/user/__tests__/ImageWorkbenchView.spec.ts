import { mount, flushPromises } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ImageWorkbenchView from '../ImageWorkbenchView.vue'

const mocks = vi.hoisted(() => ({
  editImage: vi.fn(),
  generateImage: vi.fn(),
  listKeys: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
  showWarning: vi.fn()
}))

vi.mock('@/api/keys', () => ({
  keysAPI: {
    list: mocks.listKeys
  }
}))

vi.mock('@/api/imageWorkbench', () => ({
  generateImage: mocks.generateImage,
  editImage: mocks.editImage
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: mocks.showError,
    showSuccess: mocks.showSuccess,
    showWarning: mocks.showWarning
  })
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (params?.count) return `${key}:${params.count}`
        return key
      }
    })
  }
})

function makeKey(overrides: Record<string, unknown> = {}) {
  return {
    id: 1,
    key: 'sk-image',
    name: 'Image Key',
    status: 'active',
    group: {
      name: 'OpenAI Image',
      platform: 'openai',
      allow_image_generation: true
    },
    ...overrides
  }
}

function mountView() {
  return mount(ImageWorkbenchView, {
    global: {
      stubs: {
        AppLayout: { template: '<main><slot /></main>' },
        EmptyState: {
          props: ['title', 'description', 'actionText'],
          template: '<div><slot name="icon" /><h3>{{ title }}</h3><p>{{ description }}</p><button v-if="actionText">{{ actionText }}</button><slot /></div>'
        },
        Icon: { template: '<span />' },
        Select: {
          props: ['modelValue', 'options', 'placeholder'],
          emits: ['update:modelValue'],
          template: '<select :value="modelValue ?? \'\'" @change="$emit(\'update:modelValue\', Number($event.target.value) || $event.target.value)"><option value="">{{ placeholder }}</option><option v-for="option in options" :key="String(option.value)" :value="option.value">{{ option.label }}</option></select>'
        },
        TextArea: {
          props: ['modelValue', 'label'],
          emits: ['update:modelValue'],
          template: '<label><span>{{ label }}</span><textarea :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" /></label>'
        }
      }
    }
  })
}

describe('ImageWorkbenchView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    mocks.listKeys.mockResolvedValue({ items: [] })
    mocks.generateImage.mockResolvedValue({
      images: [{ src: 'data:image/png;base64,abc', revisedPrompt: 'revised prompt' }]
    })
    mocks.editImage.mockResolvedValue({
      images: [{ src: 'data:image/png;base64,edited' }]
    })
    vi.stubGlobal('fetch', vi.fn(async () => ({ blob: async () => new Blob(['abc'], { type: 'image/png' }) })))
    Object.defineProperty(URL, 'createObjectURL', {
      configurable: true,
      value: vi.fn(() => 'blob:preview')
    })
    Object.defineProperty(URL, 'revokeObjectURL', {
      configurable: true,
      value: vi.fn()
    })
  })

  it('shows an empty state when no usable image key exists', async () => {
    mocks.listKeys.mockResolvedValue({
      items: [
        makeKey({ id: 1, status: 'inactive' }),
        makeKey({ id: 2, group: { name: 'Anthropic', platform: 'anthropic', allow_image_generation: true } }),
        makeKey({ id: 3, group: { name: 'OpenAI Text', platform: 'openai', allow_image_generation: false } })
      ]
    })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('imageWorkbench.noKeyTitle')
  })

  it('enables generation when a usable key exists and prompt has content', async () => {
    mocks.listKeys.mockResolvedValue({ items: [makeKey()] })

    const wrapper = mountView()
    await flushPromises()
    await wrapper.find('textarea').setValue('Create a product render')

    const button = wrapper.findAll('button').find((item) => item.text().includes('imageWorkbench.generateButton'))
    expect(button?.attributes('disabled')).toBeUndefined()
  })

  it('renders results after a successful generation', async () => {
    mocks.listKeys.mockResolvedValue({ items: [makeKey()] })

    const wrapper = mountView()
    await flushPromises()
    await wrapper.find('textarea').setValue('Create a product render')
    await wrapper.findAll('button').find((item) => item.text().includes('imageWorkbench.generateButton'))!.trigger('click')
    await flushPromises()

    expect(mocks.generateImage).toHaveBeenCalledWith(expect.objectContaining({ apiKey: 'sk-image', prompt: 'Create a product render' }))
    expect(wrapper.find('img[alt="Create a product render"]').exists()).toBe(true)
  })

  it('runs one-image requests concurrently for the selected count', async () => {
    mocks.listKeys.mockResolvedValue({ items: [makeKey()] })
    mocks.generateImage
      .mockResolvedValueOnce({ images: [{ src: 'data:image/png;base64,one' }] })
      .mockResolvedValueOnce({ images: [{ src: 'data:image/png;base64,two' }] })

    const wrapper = mountView()
    await flushPromises()
    await wrapper.find('textarea').setValue('Create two product renders')
    await wrapper.findAll('select')[2].setValue('2')
    await wrapper.findAll('button').find((item) => item.text().includes('imageWorkbench.generateButton'))!.trigger('click')
    await flushPromises()

    expect(mocks.generateImage).toHaveBeenCalledTimes(2)
    expect(mocks.generateImage).toHaveBeenNthCalledWith(1, expect.objectContaining({ n: 1 }))
    expect(mocks.generateImage).toHaveBeenNthCalledWith(2, expect.objectContaining({ n: 1 }))
    expect(wrapper.findAll('article')).toHaveLength(2)
  })

  it('sets a selected result as reference and switches the next request to edits', async () => {
    mocks.listKeys.mockResolvedValue({ items: [makeKey()] })

    const wrapper = mountView()
    await flushPromises()
    await wrapper.find('textarea').setValue('Create a product render')
    await wrapper.findAll('button').find((item) => item.text().includes('imageWorkbench.generateButton'))!.trigger('click')
    await flushPromises()
    await wrapper.findAll('button').find((item) => item.text().includes('imageWorkbench.setReference'))!.trigger('click')
    await flushPromises()
    await wrapper.findAll('button').find((item) => item.text().includes('imageWorkbench.editButton'))!.trigger('click')
    await flushPromises()

    expect(mocks.editImage).toHaveBeenCalledWith(expect.objectContaining({ apiKey: 'sk-image', prompt: 'Create a product render' }))
  })
})
