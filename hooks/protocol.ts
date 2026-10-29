import type { Presses, Session } from './game'

export type DinoProps = {
  saved: Session | null
  highScore: number
  resumes: number
  presses: Presses
}

export type SavedMessage = { saved: Session }

export function isSavedMessage(data: unknown): data is SavedMessage {
  if (typeof data !== 'object' || data === null || !('saved' in data)) {
    return false
  }
  const { saved } = data
  if (typeof saved !== 'object' || saved === null || !('game' in saved)) {
    return false
  }
  const { game } = saved
  return (
    typeof game === 'object' &&
    game !== null &&
    'score' in game &&
    typeof game.score === 'number' &&
    'highScore' in game &&
    typeof game.highScore === 'number'
  )
}
