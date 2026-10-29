import type { EngineInterface, Register } from 'claude-code'

import { NO_PRESSES, countPress } from './game'
import type { Presses, Session } from './game'
import { isSavedMessage } from './protocol'
import type { DinoProps } from './protocol'

const HIGH_SCORE = 'highScore'
const CONTROLS_WIDTH = 18

async function persistHighScore($: EngineInterface, highScore: number, stored: number): Promise<number> {
  if (highScore > stored) {
    await $.store.set(HIGH_SCORE, highScore)
  }
  return Math.max(highScore, stored)
}

export const register: Register = on => {
  let saved: Session | null = null
  let highScore = 0
  let storedHighScore = 0
  let resumes = 0
  let wasWorking = false
  let presses: Presses = NO_PRESSES

  on('session.start', async ($, e, next) => {
    storedHighScore = Number((await $.store.get(HIGH_SCORE)) ?? 0)
    highScore = storedHighScore
    return next(e)
  })

  on('ui.message', async ($, e, next) => {
    if (e.element !== 'dino' || !isSavedMessage(e.data)) {
      return next(e)
    }
    saved = e.data.saved
    highScore = Math.max(highScore, saved.game.highScore)
    if (saved.game.isGameOver) {
      storedHighScore = await persistHighScore($, highScore, storedHighScore)
    }
    return next(e)
  })

  on('turn.complete', async ($, e, next) => {
    storedHighScore = await persistHighScore($, highScore, storedHighScore)
    return next(e)
  })

  on('ui.render', { component: 'AbovePrompt' }, ($, e, next) => {
    const isPlayable = e.surface === 'terminal' || e.surface === 'desktop'
    if (!isPlayable || !e.props.isWorking || e.props.hasSurvey) {
      wasWorking = false
      return next(e)
    }
    if (!wasWorking) {
      wasWorking = true
      resumes += 1
    }
    const { Box, Button, Client } = $.ui.resolve(e)
    const props: DinoProps = { saved, highScore, resumes, presses }
    const onPress = (key: keyof Presses) => () => {
      presses = countPress(presses, key)
      $.ui.invalidate('ui.render')
    }

    return (
      <Box flexDirection="row">
        <Client
          key="dino"
          module="./dino.tsx"
          props={props}
          width={Math.max(0, e.props.bodyColumns - CONTROLS_WIDTH)}
          height={2}
        />
        <Box flexDirection="column" width={CONTROLS_WIDTH} paddingLeft={2}>
          <Box flexDirection="row" gap={2}>
            <Button key="jump" plain autoFocus label="␣ jump" onPress={onPress('jump')} />
            <Button key="pause" plain hotkey="p" label="pause" onPress={onPress('pause')} />
          </Box>
          <Button key="restart" plain hotkey="r" label="restart" onPress={onPress('restart')} />
        </Box>
      </Box>
    )
  })
}
