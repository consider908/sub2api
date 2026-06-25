import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick } from 'vue'
import { mount } from '@vue/test-utils'

vi.mock('@/components/icons/Icon.vue', () => ({
  default: defineComponent({
    name: 'Icon',
    template: '<span class="icon-stub" />',
  }),
}))

import BaseDialog, { __resetBaseDialogModalLockForTests } from '../BaseDialog.vue'

describe('BaseDialog', () => {
  beforeEach(() => {
    __resetBaseDialogModalLockForTests()
  })

  it('keeps body scroll lock until the last nested dialog closes', async () => {
    const outerWrapper = mount(BaseDialog, {
      props: {
        show: true,
        title: 'outer',
      },
    })
    const innerWrapper = mount(BaseDialog, {
      props: {
        show: true,
        title: 'inner',
      },
    })

    await nextTick()
    expect(document.body.classList.contains('modal-open')).toBe(true)

    await innerWrapper.setProps({ show: false })
    await nextTick()
    expect(document.body.classList.contains('modal-open')).toBe(true)

    await outerWrapper.setProps({ show: false })
    await nextTick()
    expect(document.body.classList.contains('modal-open')).toBe(false)

    innerWrapper.unmount()
    outerWrapper.unmount()
  })

  it('releases its lock on unmount without making the counter negative', async () => {
    const wrapper = mount(BaseDialog, {
      props: {
        show: true,
        title: 'dialog',
      },
    })

    await nextTick()
    expect(document.body.classList.contains('modal-open')).toBe(true)

    wrapper.unmount()
    await nextTick()
    expect(document.body.classList.contains('modal-open')).toBe(false)

    const secondWrapper = mount(BaseDialog, {
      props: {
        show: true,
        title: 'dialog-2',
      },
    })
    await nextTick()
    expect(document.body.classList.contains('modal-open')).toBe(true)

    secondWrapper.unmount()
    await nextTick()
    expect(document.body.classList.contains('modal-open')).toBe(false)
  })
})
