# Claudinosaur - Port of the Chrome dinosaur game to Claude Code

This project is a Claude Code mod that shows a small emoji-based game above the prompt box whenever Claude is working (i.e. whenever Claude is not waiting for user input, whether it is exploring, starting subagents, or writing in auto-accept mode).

## Vision for the UI of the project:

### Default Claude Code when working

----------------------------------------------------------------------------------------------------------------------------------

 ▐▛███▜▌   Claude Code v2.1.19
▝▜█████▛▘  Opus 4.5 · Claude API
  ▘▘ ▝▝    [some project]

❯ explain how this project works

● Let me explore the codebase to understand how it works.

● Read(docs/architecture.md)
  ⎿  Read 56 lines

● Search(pattern: "**/*.rs")
  ⎿  Found 57 files (ctrl+o to expand)

● Read(some_file.rs)
  ⎿  Read 561 lines

* Zigzagging… (Esc to interrupt · thought for 1s)

─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯ 
─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  -- INSERT --


----------------------------------------------------------------------------------------------------------------------------------


### With Claudinosaur

----------------------------------------------------------------------------------------------------------------------------------

 ▐▛███▜▌   Claude Code v2.1.19
▝▜█████▛▘  Opus 4.5 · Claude API
  ▘▘ ▝▝    [some project]

❯ explain how this project works

● Let me explore the codebase to understand how it works.

● Read(docs/architecture.md)
  ⎿  Read 56 lines

● Search(pattern: "**/*.rs")
  ⎿  Found 57 files (ctrl+o to expand)

● Read(some_file.rs)
  ⎿  Read 561 lines

* Zigzagging… (Esc to interrupt · thought for 1s)

    ☁️           ☁️                    ☁️                                                            [SKY LINE]
🦖                🌵                                           Score: 00060 HI: 00060 [R]estart      [GROUND LINE]
─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯ 
─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  -- INSERT --


----------------------------------------------------------------------------------------------------------------------------------

# Technical strategy

## Tech stack

A Claude Code mod: a plugin folder whose `hooks/hooks.json` names a TypeScript hooks module. The repository root is the plugin folder.

```bash
claude --plugin-dir .          # run Claude Code with the mod loaded
claude plugin validate .       # what the engine sees and would refuse
claude plugin test .           # unit and band tests
```

The mod API is early access and moves between releases. Its declarations are written by the engine into `.claude-plugin/types/` (gitignored) whenever a session hot-reloads the folder; grep them for an event or element name rather than guessing.

## Application lifecycle

- `hooks/register.tsx` hooks `ui.render` on the `AbovePrompt` band. While `e.props.isWorking` is true it draws the game; otherwise it calls `next(e)` and the band is the engine's again.
- The game itself is `hooks/dino.tsx`, a `Client` surface module: it keeps the game state locally, ticks on `surface.every(50)` and posts a snapshot of the session to the hooks module every 4 frames and on every key.
- Each idle→working transition bumps a `resumes` counter in the `Client`'s props. On a new value the game pauses behind the 3 second countdown, whether the engine kept the instance or mounted a new one from the last snapshot.
- The high score lives in `$.store`, so it survives across sessions.

## Keyboard input

- A `Client` only receives keys after a click, and clicks are only reported in fullscreen mode.
- Keyboard play therefore goes through Buttons drawn by `register.tsx` beside the game: `ctrl+x tab` focuses the band, Space/Enter press the auto-focused jump Button, `p`/`r` are hotkeys, `Esc` returns to the prompt.
- Button presses reach the `Client` as per-key counters in its props (`Presses`); the game replays the presses it has not seen.
- Buttons drawn inside a `Client` are not part of the band's focus ring.
- While the band is not focused every key goes to the prompt, so the person can keep typing while the game runs.

## Rendering

- Each line is laid out as one cell per terminal column: a wide glyph fills its cell and an empty continuation cell, so a line is always exactly as wide as the `Client`.
- `☁️` (U+2601 + VS16) is drawn one column wide by some terminals while the engine counts two; only use glyphs that are always wide (`⛅`, `🦕`, `🌵`, `🦅`, `💀`).

## Architecture

Follow a "functional core, imperative shell" approach. Code related to physics, scoring, rendering of the two lines of the game should be isolated from the mod-specific logic.
The core part should be pure functions covered by unit tests.

### Code structure

```
├── .claude-plugin/plugin.json
├── hooks/
│   ├── hooks.json      # names register.tsx as the hooks module
│   ├── register.tsx    # imperative shell: AbovePrompt band, control Buttons, snapshot and high score
│   ├── dino.tsx        # imperative shell: Client surface module, frame clock, local state
│   ├── game.ts         # pure game logic: physics, scoring, countdown, presses
│   ├── render.ts       # pure rendering of the sky and ground lines
│   └── protocol.ts     # props and messages between register.tsx and dino.tsx
├── tests/
│   ├── game.test.ts    # unit tests of the pure core
│   └── band.test.tsx   # the band through the engine's test kit, on terminal and desktop
└── scripts/
    └── capture_mod.sh  # frames of real Claude Code with the mod, via tmux
```

## Integration testing

`claude plugin test .` mounts the band through the engine itself (`$.ui.mount`) and drives the `Client` with `ui.key`, `ui.press` and `ui.advance` on its frame clock.

`scripts/capture_mod.sh` drives real Claude Code in a detached 120x32 tmux session and saves one text frame per step of the lifecycle: idle, working, jump, pause, idle again, resume countdown. It sends real prompts, so it costs API usage. Frames are saved to `.claude/screenshots/mod_<timestamp>/`.

## Coding style

Comments should be exceptional. Only write comments when it is unclear WHY a piece of code was needed. If you need a comment to explain what your code does, then your code is not explicit enough and that is a problem.

## Git

Follow conventional commit style. Commits should only have a message, not need for a commit body. Make small, bite-sized commits, do not commit a whole feature at once.

# Gameplay

## Controls

- <ctrl+x tab> to focus the game, <Esc> to hand the keyboard back to the prompt
- <space> to jump
- <R> to restart
- <P> to pause

## Automatic pausing

When Claude Code is done thinking, the game should be paused, and the app should store current state. When Claude Code resume working, the game should re-appear and there should be a 3 second countdown before it restarts
