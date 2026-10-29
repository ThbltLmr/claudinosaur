import { BIRD, CACTUS, CLOUD, DEAD, DINO } from './game'
import type { Session, State } from './game'

export const MIN_WIDTH = 20
export const FOCUS_HINT = 'click here, then space to jump'

// A wide glyph fills its own cell and an empty one after it, so the joined
// line is exactly `width` terminal columns.
const CONTINUATION = ''

type Cells = string[]

function blank(width: number): Cells {
  return Array.from({ length: width }, () => ' ')
}

function clearCell(cells: Cells, i: number): void {
  if (cells[i] === CONTINUATION && i > 0) {
    cells[i - 1] = ' '
  }
  if (cells[i + 1] === CONTINUATION) {
    cells[i + 1] = ' '
  }
}

function putEmoji(cells: Cells, x: number, emoji: string): void {
  const pos = Math.floor(x)
  if (pos < 0 || pos + 1 >= cells.length) {
    return
  }
  clearCell(cells, pos)
  clearCell(cells, pos + 1)
  cells[pos] = emoji
  cells[pos + 1] = CONTINUATION
}

function putText(cells: Cells, pos: number, text: string): void {
  Array.from(text).forEach((char, i) => {
    const at = pos + i
    if (at >= 0 && at < cells.length) {
      clearCell(cells, at)
      cells[at] = char
    }
  })
}

export function formatScore(s: State): string {
  const pad = (n: number) => String(n).padStart(5, '0')
  const base = `Score: ${pad(s.score)} HI: ${pad(s.highScore)}`
  return s.isGameOver ? `${base} [R]estart` : base
}

export function formatCountdown(seconds: number): string {
  return `▶ ${Math.ceil(seconds)} ◀`
}

function skyLine(session: Session, width: number): string {
  const { game } = session
  const cells = blank(width)
  game.clouds.forEach(x => putEmoji(cells, x, CLOUD))
  game.birds.forEach(x => putEmoji(cells, x, BIRD))
  if (game.isInAir && !game.isGameOver) {
    putEmoji(cells, 0, DINO)
  }
  if (session.countdown > 0) {
    const label = formatCountdown(session.countdown)
    putText(cells, Math.floor((width - label.length) / 2), label)
  } else if (!session.hasPlayed) {
    putText(cells, width - FOCUS_HINT.length, FOCUS_HINT)
  }
  return cells.join('')
}

function groundLine(game: State, width: number): string {
  const cells = blank(width)
  if (!game.isInAir) {
    putEmoji(cells, 0, game.isGameOver ? DEAD : DINO)
  }
  game.obstacles.forEach(x => putEmoji(cells, x, CACTUS))
  const score = formatScore(game)
  putText(cells, width - score.length, score)
  return cells.join('')
}

export function renderLines(session: Session, width: number): [sky: string, ground: string] {
  if (width < MIN_WIDTH) {
    return [CLOUD, DINO]
  }
  return [skyLine(session, width), groundLine(session.game, width)]
}
