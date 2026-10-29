export const JUMP_DURATION = 0.4
export const BASE_OBSTACLE_SPEED = 40
export const SPEED_INCREASE_RATE = 2
export const BASE_SPAWN_INTERVAL = 2
export const MIN_SPAWN_INTERVAL = 0.8
export const SPAWN_DECREASE_RATE = 0.1
export const DINO_HITBOX_END = 2
export const GAME_OVER_DELAY = 2

export const CLOUD_SPEED = 15
export const MIN_CLOUD_SPACING = 10

export const BIRD_SCORE_THRESHOLD = 200
export const BIRD_SPAWN_INTERVAL = 3
export const BIRD_SAFE_ZONE = 5

export const RESUME_COUNTDOWN = 3

export const DINO = '🦕'
export const DEAD = '💀'
export const CACTUS = '🌵'
export const CLOUD = '⛅'
export const BIRD = '🦅'

export type State = {
  isInAir: boolean
  jumpTimeLeft: number
  obstacles: number[]
  score: number
  highScore: number
  isGameOver: boolean
  gameOverTime: number
  isPaused: boolean
  elapsedTime: number
  timeSinceSpawn: number
  clouds: number[]
  birds: number[]
  timeSinceBird: number
}

export type Session = {
  game: State
  countdown: number
  hasPlayed: boolean
}

export type Key = 'jump' | 'restart' | 'pause'

export type Presses = Record<Key, number>

export const NO_PRESSES: Presses = { jump: 0, restart: 0, pause: 0 }

const KEYS: Key[] = ['jump', 'restart', 'pause']

export function newState(highScore = 0): State {
  return {
    isInAir: false,
    jumpTimeLeft: 0,
    obstacles: [],
    score: 0,
    highScore,
    isGameOver: false,
    gameOverTime: 0,
    isPaused: false,
    elapsedTime: 0,
    timeSinceSpawn: 0,
    clouds: [4, 20, 45],
    birds: [],
    timeSinceBird: 0,
  }
}

export function tick(s: State, dt: number, width: number): State {
  if (s.isPaused) {
    return s
  }

  if (s.isGameOver) {
    const gameOverTime = s.gameOverTime + dt
    return gameOverTime >= GAME_OVER_DELAY ? restart(s) : { ...s, gameOverTime }
  }

  let next: State = {
    ...s,
    elapsedTime: s.elapsedTime + dt,
    timeSinceSpawn: s.timeSinceSpawn + dt,
  }
  next = updateJump(next, dt)
  next = updateObstacles(next, dt)
  next = updateClouds(next, dt)
  next = updateBirds(next, dt)
  next = spawnObstacle(next, width)
  next = spawnCloud(next, width)
  next = spawnBird(next, width)

  if (hitsCactus(next) || hitsBird(next)) {
    return {
      ...next,
      isGameOver: true,
      gameOverTime: 0,
      highScore: Math.max(next.highScore, next.score),
    }
  }

  const score = next.score + 1
  return { ...next, score, highScore: Math.max(next.highScore, score) }
}

export function jump(s: State): State {
  if (s.isInAir || s.isGameOver || s.isPaused) {
    return s
  }
  return { ...s, isInAir: true, jumpTimeLeft: JUMP_DURATION }
}

export function restart(s: State): State {
  return newState(s.highScore)
}

export function togglePause(s: State): State {
  return s.isGameOver ? s : { ...s, isPaused: !s.isPaused }
}

function updateJump(s: State, dt: number): State {
  if (!s.isInAir) {
    return s
  }
  const jumpTimeLeft = s.jumpTimeLeft - dt
  return jumpTimeLeft <= 0
    ? { ...s, isInAir: false, jumpTimeLeft: 0 }
    : { ...s, jumpTimeLeft }
}

function moveLeft(positions: number[], distance: number): number[] {
  return positions.map(x => x - distance).filter(x => x > -2)
}

function updateObstacles(s: State, dt: number): State {
  return { ...s, obstacles: moveLeft(s.obstacles, currentSpeed(s.elapsedTime) * dt) }
}

function updateClouds(s: State, dt: number): State {
  return { ...s, clouds: moveLeft(s.clouds, CLOUD_SPEED * dt) }
}

function updateBirds(s: State, dt: number): State {
  return {
    ...s,
    timeSinceBird: s.timeSinceBird + dt,
    birds: moveLeft(s.birds, currentSpeed(s.elapsedTime) * dt),
  }
}

function spawnObstacle(s: State, width: number): State {
  if (s.timeSinceSpawn < spawnInterval(s.score)) {
    return s
  }
  return { ...s, timeSinceSpawn: 0, obstacles: [...s.obstacles, width] }
}

function spawnCloud(s: State, width: number): State {
  const rightmost = s.clouds[s.clouds.length - 1]
  if (rightmost !== undefined && width - rightmost < MIN_CLOUD_SPACING) {
    return s
  }
  return { ...s, clouds: [...s.clouds, width] }
}

function spawnBird(s: State, width: number): State {
  if (s.score < BIRD_SCORE_THRESHOLD || s.timeSinceBird < BIRD_SPAWN_INTERVAL) {
    return s
  }
  if (hasConflict(s.clouds, width) || hasConflict(s.obstacles, width)) {
    return s
  }
  return { ...s, timeSinceBird: 0, birds: [...s.birds, width] }
}

function isAtDino(x: number): boolean {
  return x >= 0 && x <= DINO_HITBOX_END
}

function hitsCactus(s: State): boolean {
  return !s.isInAir && s.obstacles.some(isAtDino)
}

function hitsBird(s: State): boolean {
  return s.isInAir && s.birds.some(isAtDino)
}

function hasConflict(positions: number[], x: number): boolean {
  return positions.some(pos => pos >= x - BIRD_SAFE_ZONE && pos <= x + BIRD_SAFE_ZONE)
}

export function currentSpeed(elapsed: number): number {
  return BASE_OBSTACLE_SPEED + (elapsed / 10) * SPEED_INCREASE_RATE
}

export function spawnInterval(score: number): number {
  return Math.max(MIN_SPAWN_INTERVAL, BASE_SPAWN_INTERVAL - (score / 100) * SPAWN_DECREASE_RATE)
}

export function startSession(saved: Session | null, highScore: number): Session {
  if (saved === null) {
    return { game: newState(highScore), countdown: 0, hasPlayed: false }
  }
  const game = { ...saved.game, highScore: Math.max(saved.game.highScore, highScore) }
  return resume({ ...saved, game })
}

export function resume(session: Session): Session {
  if (session.game.isGameOver) {
    return { ...session, countdown: 0 }
  }
  return { ...session, game: { ...session.game, isPaused: true }, countdown: RESUME_COUNTDOWN }
}

export function advance(session: Session, dt: number, width: number): Session {
  if (session.countdown <= 0) {
    return { ...session, game: tick(session.game, dt, width) }
  }
  const countdown = Math.max(0, session.countdown - dt)
  if (countdown > 0) {
    return { ...session, countdown }
  }
  return { ...session, countdown, game: { ...session.game, isPaused: false } }
}

export function press(session: Session, key: Key): Session {
  if (session.countdown > 0) {
    return session
  }
  const act = { jump, restart, pause: togglePause }[key]
  return { ...session, hasPlayed: true, game: act(session.game) }
}

export function countPress(presses: Presses, key: Key): Presses {
  return { ...presses, [key]: presses[key] + 1 }
}

export function isSamePresses(a: Presses, b: Presses): boolean {
  return KEYS.every(key => a[key] === b[key])
}

export function catchUp(session: Session, seen: Presses, latest: Presses): Session {
  return KEYS.reduce((current, key) => {
    const missed = Math.max(0, latest[key] - seen[key])
    return Array.from({ length: missed }).reduce<Session>(s => press(s, key), current)
  }, session)
}

export function keyFor(typed: string): Key | undefined {
  if (typed === ' ' || typed === 'space') {
    return 'jump'
  }
  const lower = typed.toLowerCase()
  if (lower === 'r') {
    return 'restart'
  }
  if (lower === 'p') {
    return 'pause'
  }
  return undefined
}
