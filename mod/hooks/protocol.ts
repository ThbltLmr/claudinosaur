import type { State } from './game'

export type DinoProps = {
  saved: State | null
  highScore: number
}

export type SavedMessage = { saved: State }

export function isSavedMessage(data: unknown): data is SavedMessage {
  if (typeof data !== 'object' || data === null || !('saved' in data)) {
    return false
  }
  const { saved } = data
  return (
    typeof saved === 'object' &&
    saved !== null &&
    'score' in saved &&
    typeof saved.score === 'number' &&
    'highScore' in saved &&
    typeof saved.highScore === 'number'
  )
}
