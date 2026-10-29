import { expect, mock, test } from 'claude-code/testing'

import { FOCUS_HINT } from '../hooks/render'

const WIDTH = 80

const BAND = {
  plugin: 'claudinosaur',
  component: 'AbovePrompt',
  props: {
    hasSurvey: false,
    isWorking: true,
    maxRows: 12,
    bodyColumns: WIDTH,
    scroll: { offset: 0, bodyRows: 11 },
    view: {},
  },
} as const

const IDLE = { ...BAND.props, isWorking: false }

for (const surface of ['terminal', 'desktop'] as const) {
  test(`${surface}: the band stays the engine's while Claude is idle`, async ($, on) => {
    on('ui.render', ($, e) => {
      const { Text } = $.ui.resolve(e)
      return <Text>engine</Text>
    })

    const ui = await $.ui.mount({ ...BAND, props: IDLE, surface })

    expect(await ui.find({ type: 'Client' })).toBeUndefined()
    expect((await ui.find({ type: 'Text' }))?.text).toBe('engine')
    await ui.unmount()
  })

  test(`${surface}: the game runs while Claude works and jumps on space`, async $ => {
    const ui = await $.ui.mount({ ...BAND, surface })
    await ui.resize({ columns: WIDTH, rows: 2, in: 'dino' })

    expect(await ui.find({ in: 'dino', text: new RegExp(FOCUS_HINT) })).toBeDefined()

    await ui.advance(500)
    expect(await ui.find({ in: 'dino', text: /Score: 00010 HI: 00010$/ })).toBeDefined()

    await ui.key({ key: ' ', in: 'dino' })
    const [sky, ground] = await ui.findAll({ in: 'dino', type: 'Text' })
    expect(sky?.text.startsWith('🦕')).toBe(true)
    expect(ground?.text.startsWith('🦕')).toBe(false)
    expect(sky?.text).not.toContain(FOCUS_HINT)
    await ui.unmount()
  })

  test(`${surface}: going idle pauses the game and working again resumes it after a countdown`, async ($, on) => {
    on('ui.render', ($, e) => {
      const { Text } = $.ui.resolve(e)
      return <Text>engine</Text>
    })

    const ui = await $.ui.mount({ ...BAND, surface })
    await ui.resize({ columns: WIDTH, rows: 2, in: 'dino' })
    await ui.advance(1000)

    await ui.redraw(IDLE)
    expect(await ui.find({ type: 'Client' })).toBeUndefined()

    await ui.redraw(BAND.props)
    await ui.resize({ columns: WIDTH, rows: 2, in: 'dino' })
    expect(await ui.find({ in: 'dino', text: /▶ 3 ◀/ })).toBeDefined()
    expect(await ui.find({ in: 'dino', text: /Score: 00020/ })).toBeDefined()

    await ui.advance(2900)
    expect(await ui.find({ in: 'dino', text: /Score: 00020/ })).toBeDefined()

    await ui.advance(200)
    expect(await ui.find({ in: 'dino', text: /Score: 0002[12]/ })).toBeDefined()
    expect(await ui.find({ in: 'dino', text: /▶/ })).toBeUndefined()
    await ui.unmount()
  })
}

test('the high score is kept across sessions', async ($, on) => {
  mock.store(on, { highScore: 300 })
  on('session.start', (_, e) => ({ cwd: e.cwd }))
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })

  const ui = await $.ui.mount({ ...BAND, surface: 'terminal' })
  await ui.resize({ columns: WIDTH, rows: 2, in: 'dino' })
  expect(await ui.find({ in: 'dino', text: /HI: 00300$/ })).toBeDefined()

  await ui.unmount()
})

test('a new best score is stored when the dino dies', async ($, on) => {
  const store: Record<string, unknown> = { highScore: 10 }
  on('store.get', (_, e) => ({ value: store[e.key] }))
  on('store.set', (_, e) => {
    store[e.key] = e.value
    return { value: undefined }
  })
  on('session.start', (_, e) => ({ cwd: e.cwd }))
  await $.session.start({ cwd: '/', surface: 'terminal', isInteractive: true })

  const ui = await $.ui.mount({ ...BAND, surface: 'terminal' })
  await ui.resize({ columns: WIDTH, rows: 2, in: 'dino' })
  await ui.advance(5000)

  expect(await ui.find({ in: 'dino', text: /💀/ })).toBeDefined()
  expect(Number(store.highScore)).toBeGreaterThan(10)
  await ui.unmount()
})
