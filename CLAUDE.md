# Claudinosaur - Port of the Chrome dinosaur game to Claude Code

This project is a PTY wrapper around Claude Code that detects when Claude is working (i.e. whenever Claude is not waiting for user input, whether it is exploring, starting subagents, or writing in auto-accept mode), and insert a small emoji-based game above the prompt box.

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

Go with bubbletea

## Application lifecycle

The wrapper needs to have two modes:
- passthrough (does nothing, passes all inputs to Claude)
- game active (overlay game content on the terminal)

State detection uses spinner characters (`✢✶✻✸✹✺✷`) with a 500ms timeout.

## Technical Learnings (from Step 3)

### TUI Rendering Model

Claude Code is a cursor-based TUI. It uses ANSI escape sequences to position the cursor and redraw specific screen regions:
- `\x1b[H` - move cursor to home
- `\x1b[<row>;<col>H` - move to specific position
- `\x1b[2K` - clear current line
- `\x1b[s` / `\x1b[u` - save/restore cursor position

CC constantly redraws its status line (spinner animation) using these sequences. It assumes it knows the screen layout.

### Why Inline Injection Fails

When we inject bytes directly into the PTY stream:
1. Our content appears on screen
2. CC doesn't know about it - its internal state doesn't match reality
3. CC's next cursor positioning writes over our content or in wrong locations
4. Visual artifacts everywhere

### Overlay Approach (Current Strategy)

Instead of injecting into the stream, we use an **overlay** technique:
1. Pass all CC output through **unchanged**
2. Track cursor position by parsing ANSI escape sequences (`\x1b[row;colH`, etc.)
3. When spinner character detected, record which row it's on (`lastSpinnerRow`)
4. After a "quiet period" (10ms of no new bytes), CC has finished its frame
5. Render two game lines at `spinnerRow + 2` (sky) and `spinnerRow + 3` (ground)
6. Use ANSI save/restore cursor (`\x1b[s` / `\x1b[u`) to avoid disrupting CC's cursor
7. Clear terminal on startup for accurate cursor tracking
8. Clear previous overlay position when spinner row changes
9. On transition to idle, restore the input prompt area (dash line + prompt)

This creates a stable two-line overlay without breaking CC's assumptions.

Shutdown:
- the application should be stopped with <ctrl-C>. It is okay to overwrite the Claude Code <ctrl-C> behaviour, since users can alternatively use <Esc> to interrupt Claude Code actions anyway.
- pressing <ctrl-C> in Claudinosaur should shutdown the underlying Claude Code instance, shut down the wrapper, and return to default terminal behaviour.

## Architecture

Follow a "functional core, imperative shell" approach. Code related to physics, scoring, rendering of the two lines of the game should be isolated from the bubbletea specific logic.
The core part should be pure functions covered by unit tests.

### Current Code Structure

```
├── main.go           # Entry point, PTY setup, signal handling, bubbletea program
├── state/
│   └── detector.go   # Spinner-based state detection (idle/working)
├── inject/
│   ├── transformer.go  # Mode enum and passthrough transform
│   ├── overlay.go      # ANSI overlay rendering (RenderMultiLineOverlay, ClearMultipleRows)
│   └── cursor.go       # ANSI cursor position tracking, spinner row detection
├── ui/
│   └── model.go      # Bubbletea model, quiet period detection, game line generation
└── game/             # (To be created in Step 5) Pure game logic
```

## Integration testing - To be investigated

Open question: have a "--test" mode where the whole rendered screen is compared to snapshots instead of being passed to Claude Code to prevent regressions?

## Coding style

Comments should be exceptional. Only write comments when it is unclear WHY a piece of code was needed. If you need a comment to explain what your code does, then your code is not explicit enough and that is a problem.

## Git

Follow conventional commit style. Commits should only have a message, not need for a commit body. Make small, bite-sized commits, do not commit a whole feature at once.

# Gameplay

## Controls

- <space> to jump
- <R> to restart
- <P> to pause

## Automatic pausing

When Claude Code is done thinking, the game should be paused, and the app should store current state. When Claude Code resume working, the game should re-appear and there should be a 3 second countdown before it restarts
