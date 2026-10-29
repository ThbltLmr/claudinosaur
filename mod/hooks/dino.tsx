import type { ClientModule } from 'claude-code'

import { advance, keyFor, press, startSession } from './game'
import type { Session } from './game'
import { renderLines } from './render'
import type { DinoProps } from './protocol'

export const FRAME_MS = 50
const FRAMES_PER_SAVE = 4

type Local = { session: Session; frame: number }

const Dino: ClientModule<DinoProps, Local> = (props, surface) => {
  const { Box, Text } = surface.elements

  const save = (session: Session) => surface.post({ saved: session.game })

  const apply = (change: (session: Session) => Session, isSaveForced = false) => {
    const current = surface.state
    if (current === undefined) {
      return
    }
    const session = change(current.session)
    const frame = current.frame + 1
    surface.setState({ session, frame })
    const hasJustLost = session.game.isGameOver && !current.session.game.isGameOver
    if (isSaveForced || hasJustLost || frame % FRAMES_PER_SAVE === 0) {
      save(session)
    }
  }

  if (surface.state === undefined) {
    surface.setState({ session: startSession(props.saved, props.highScore), frame: 0 })
    surface.every(FRAME_MS, () => apply(s => advance(s, FRAME_MS / 1000, surface.columns)))
    surface.onKey(event => {
      const key = keyFor(event.key)
      if (key !== undefined && !event.ctrl && !event.meta) {
        apply(s => press(s, key), true)
      }
    })
  }

  const session = surface.state?.session ?? startSession(props.saved, props.highScore)
  const [sky, ground] = renderLines(session, surface.columns)

  return (
    <Box flexDirection="column">
      <Text wrap="truncate">{sky}</Text>
      <Text wrap="truncate">{ground}</Text>
    </Box>
  )
}

export default Dino
