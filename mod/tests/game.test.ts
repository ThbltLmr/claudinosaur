import { describe, expect, test } from 'claude-code/testing'

import {
  BIRD_SCORE_THRESHOLD,
  DINO_HITBOX_END,
  GAME_OVER_DELAY,
  JUMP_DURATION,
  MIN_SPAWN_INTERVAL,
  RESUME_COUNTDOWN,
  advance,
  currentSpeed,
  jump,
  keyFor,
  newState,
  press,
  restart,
  resume,
  spawnInterval,
  startSession,
  tick,
  togglePause,
} from '../hooks/game'
import { FOCUS_HINT, renderLines } from '../hooks/render'

const WIDTH = 80
const DT = 0.05

function columns(line: string): number {
  return Array.from(line).reduce((sum, char) => {
    if (char === '️') {
      return sum
    }
    return sum + ((char.codePointAt(0) ?? 0) > 0x2000 && char !== '▶' && char !== '◀' ? 2 : 1)
  }, 0)
}

describe('tick', () => {
  test('scores one point per tick', () => {
    expect(tick(newState(), DT, WIDTH).score).toBe(1)
  })

  test('spawns a cactus at the right edge once the interval passed', () => {
    const s = tick({ ...newState(), timeSinceSpawn: 2 }, DT, WIDTH)
    expect(s.obstacles).toEqual([WIDTH])
  })

  test('ends the game when a cactus reaches a grounded dino', () => {
    const s = tick({ ...newState(), obstacles: [DINO_HITBOX_END + 0.5], score: 10 }, DT, WIDTH)
    expect(s.isGameOver).toBe(true)
    expect(s.highScore).toBe(10)
  })

  test('lets a jumping dino clear a cactus', () => {
    const s = tick(jump({ ...newState(), obstacles: [DINO_HITBOX_END + 0.5] }), DT, WIDTH)
    expect(s.isGameOver).toBe(false)
  })

  test('ends the game when a bird reaches a jumping dino', () => {
    const s = tick(jump({ ...newState(), birds: [DINO_HITBOX_END + 0.5] }), DT, WIDTH)
    expect(s.isGameOver).toBe(true)
  })

  test('lands the dino after the jump duration', () => {
    let s = jump(newState())
    for (let t = 0; t < JUMP_DURATION + DT; t += DT) {
      s = tick(s, DT, WIDTH)
    }
    expect(s.isInAir).toBe(false)
  })

  test('restarts on its own after the game over delay, keeping the high score', () => {
    const over = { ...newState(), isGameOver: true, score: 50, highScore: 50 }
    const s = tick(over, GAME_OVER_DELAY, WIDTH)
    expect(s.isGameOver).toBe(false)
    expect(s.score).toBe(0)
    expect(s.highScore).toBe(50)
  })

  test('does nothing while paused', () => {
    const paused = togglePause(newState())
    expect(tick(paused, DT, WIDTH)).toBe(paused)
  })

  test('only spawns birds past the score threshold', () => {
    const early = { ...newState(), clouds: [WIDTH - 9], timeSinceBird: 10, score: BIRD_SCORE_THRESHOLD - 10 }
    const late = { ...early, score: BIRD_SCORE_THRESHOLD }
    expect(tick(early, DT, WIDTH).birds).toEqual([])
    expect(tick(late, DT, WIDTH).birds).toEqual([WIDTH])
  })
})

describe('difficulty', () => {
  test('speeds up over time', () => {
    expect(currentSpeed(100)).toBeGreaterThan(currentSpeed(0))
  })

  test('spawns cacti more often as the score grows, down to a floor', () => {
    expect(spawnInterval(500)).toBeLessThan(spawnInterval(0))
    expect(spawnInterval(100_000)).toBe(MIN_SPAWN_INTERVAL)
  })
})

describe('controls', () => {
  test('cannot jump twice in a row', () => {
    const once = jump(newState())
    expect(jump(once)).toBe(once)
  })

  test('restart keeps only the high score', () => {
    expect(restart({ ...newState(), score: 99, highScore: 120 })).toEqual(newState(120))
  })

  test('maps the keys the game listens to', () => {
    expect(keyFor(' ')).toBe('jump')
    expect(keyFor('space')).toBe('jump')
    expect(keyFor('R')).toBe('restart')
    expect(keyFor('p')).toBe('pause')
    expect(keyFor('x')).toBeUndefined()
  })
})

describe('session', () => {
  test('a first session starts right away with the stored high score', () => {
    const session = startSession(null, 300)
    expect(session.countdown).toBe(0)
    expect(session.game.highScore).toBe(300)
    expect(session.hasPlayed).toBe(false)
  })

  test('resuming a saved game pauses it behind a countdown', () => {
    const saved = { game: { ...newState(), score: 42 }, countdown: 0, hasPlayed: true }
    const session = startSession(saved, 0)
    expect(session.countdown).toBe(RESUME_COUNTDOWN)
    expect(session.game.isPaused).toBe(true)
    expect(session.game.score).toBe(42)
  })

  test('the countdown unpauses the game when it runs out', () => {
    let session = resume(startSession(null, 0))
    session = advance(session, RESUME_COUNTDOWN, WIDTH)
    expect(session.countdown).toBe(0)
    expect(session.game.isPaused).toBe(false)
    expect(advance(session, DT, WIDTH).game.score).toBe(1)
  })

  test('resuming keeps a lost game on its game over screen', () => {
    const lost = { game: { ...newState(), isGameOver: true }, countdown: 0, hasPlayed: true }
    expect(resume(lost)).toEqual(lost)
  })

  test('keys are ignored during the countdown', () => {
    const session = resume(startSession(null, 0))
    expect(press(session, 'jump')).toBe(session)
  })
})

describe('renderLines', () => {
  test('draws the dino on the ground and the score on the right', () => {
    const [, ground] = renderLines(startSession(null, 60), WIDTH)
    expect(ground.startsWith('🦕')).toBe(true)
    expect(ground.endsWith('Score: 00000 HI: 00060')).toBe(true)
  })

  test('moves the dino to the sky line while it jumps', () => {
    const session = press(startSession(null, 0), 'jump')
    const [sky, ground] = renderLines(session, WIDTH)
    expect(sky.startsWith('🦕')).toBe(true)
    expect(ground.startsWith('🦕')).toBe(false)
  })

  test('shows a skull over the cactus and the restart hint on game over', () => {
    const game = { ...newState(), isGameOver: true, obstacles: [1.4] }
    const [, ground] = renderLines({ game, countdown: 0, hasPlayed: true }, WIDTH)
    expect(ground.startsWith('💀')).toBe(true)
    expect(ground.endsWith('[R]estart')).toBe(true)
  })

  test('explains how to play until the first key', () => {
    const fresh = startSession(null, 0)
    expect(renderLines(fresh, WIDTH)[0]).toContain(FOCUS_HINT)
    expect(renderLines(press(fresh, 'jump'), WIDTH)[0]).not.toContain(FOCUS_HINT)
  })

  test('shows the countdown while resuming', () => {
    const [sky] = renderLines(resume(startSession(null, 0)), WIDTH)
    expect(sky).toContain('▶ 3 ◀')
  })

  test('keeps every line exactly as wide as the band', () => {
    const game = { ...newState(), obstacles: [10, 77.5], clouds: [3, 30.2, 61], birds: [45] }
    const session = { game: jump(game), countdown: 0, hasPlayed: true }
    for (const line of renderLines(session, WIDTH)) {
      expect(columns(line)).toBe(WIDTH)
    }
  })

  test('falls back to two glyphs in a narrow band', () => {
    expect(renderLines(startSession(null, 0), 10)).toEqual(['☁️', '🦕'])
  })
})
