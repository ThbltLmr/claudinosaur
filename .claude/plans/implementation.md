# Claudinosaur Implementation Plan

## Overview

A PTY wrapper around Claude Code that displays an emoji-based dinosaur game while Claude is working.

**Steps should only be marked as complete after they have been tested by the user and the user has validated all criteria**

## Steps

### Step 1: PTY Passthrough + Shutdown ✓ COMPLETE

**Goal:** Create a minimal Go program that spawns Claude Code in a PTY, passes all I/O through transparently, and shuts down cleanly on Ctrl+C.

**Files to create:**
- `go.mod` - module definition
- `main.go` - entry point, PTY setup, signal handling

**Note:** No bubbletea in this step. Pure PTY + goroutines. Bubbletea introduced in step 3.

**Technical approach:**

1. Initialize a Go module (`github.com/thibault/claudinosaur`)
2. Use `github.com/creack/pty` for PTY management
3. Spawn `claude` as a subprocess attached to the PTY
4. Set up goroutines for bidirectional I/O:
   - stdin → PTY (user input to Claude)
   - PTY → stdout (Claude output to terminal)
5. Put the terminal in raw mode so keypresses are passed through immediately
6. Handle Ctrl+C (SIGINT):
   - Send SIGINT to the child process
   - Wait for child to exit
   - Restore terminal state
   - Exit wrapper
7. Forward SIGWINCH (terminal resize) to the child process

**Acceptance criteria:**
- [x] Running `go run .` launches Claude Code
- [x] All Claude Code functionality works normally (typing, responses, tool use)
- [x] Ctrl+C cleanly exits both Claude Code and the wrapper
- [x] Terminal state is restored after exit (no stuck raw mode)
- [x] Terminal resizing works correctly

**Dependencies:**
```
github.com/creack/pty
golang.org/x/term
```

---

### Step 2: State Detection ✓ COMPLETE

**Goal:** Parse Claude Code output to detect spinner characters, determine when Claude is "working" vs "idle". Log state changes for debugging.

**Files to modify/create:**
- `main.go` - add output parsing in the PTY → stdout goroutine
- `state/detector.go` - state detection logic (pure functions)
- `state/detector_test.go` - unit tests for detection logic

**Spinner characters:** `✢`, `✶`, `✻`, `✸`, `✹`, `✺`, `✷`

**Technical approach:**

1. Create a `Detector` struct that:
   - Receives chunks of output
   - Scans for any of the 7 spinner characters
   - Tracks current state (`idle` | `working`)
   - Tracks timestamp of last spinner seen
2. In `main.go`, wrap the PTY output handler to feed bytes through the detector before passing to stdout
3. Use a timeout (e.g., 500ms without spinner = transition to idle)
4. Log state transitions to a debug file (`~/.claudinosaur/debug.log`)
5. Add a `--dino-debug` flag to enable/disable state logging

**Key considerations:**
- Output arrives in chunks, but spinner chars are single Unicode codepoints → simpler than pattern matching
- Spinner disappears when Claude is done → absence of spinner for N milliseconds = idle
- Must not slow down or corrupt the passthrough (detector is read-only observer)

**Acceptance criteria:**
- [x] State changes are logged when Claude starts/stops working
- [x] Detection works across different Claude actions (thinking, tool use, subagents)
- [x] PTY passthrough remains fully functional
- [x] Unit tests cover spinner character detection and timeout logic

**Implementation notes:**
- Timeout set to 500ms
- Thread-safe with sync.Mutex
- Uses io.TeeReader for non-intrusive observation

---

### Step 3: Overlay Rendering ✓ COMPLETE

**Goal:** Display game lines when Claude is working without breaking Claude Code's UI.

**Files created/modified:**
- `main.go` - integrated bubbletea, channel-based PTY forwarding, terminal clear on startup
- `ui/model.go` - bubbletea model with quiet period detection, cursor tracking integration
- `inject/transformer.go` - simplified to pure passthrough
- `inject/overlay.go` - ANSI overlay rendering functions
- `inject/cursor.go` - ANSI cursor position tracking and spinner row detection

**What we learned:**

1. **Dash separator not in stream**: CC uses ANSI cursor positioning, not literal dash characters
2. **Inline injection breaks TUI**: CC assumes it controls the screen layout; injecting bytes causes visual artifacts
3. **Cursor tracking is essential**: Must parse ANSI sequences to know where spinner is rendered

**Implementation (overlay rendering with cursor tracking):**

1. Clear terminal on startup for accurate cursor tracking
2. Pass all CC output through **unchanged**
3. Track cursor position by parsing ANSI sequences (`\x1b[row;colH`, movements)
4. When spinner character detected, record `lastSpinnerRow`
5. Detect "quiet periods" (~16ms of no new bytes) = CC finished rendering frame
6. After quiet period, append ANSI sequences to draw overlay at `spinnerRow + 1`:
   - `\x1b[s` - save cursor
   - `\x1b[<row>;1H` - move to target row
   - `\x1b[2K` - clear line
   - Write game content
   - `\x1b[u` - restore cursor
7. Clear previous overlay position when spinner row changes
8. Clear overlay when transitioning to idle state

**Acceptance criteria:**
- [x] Game lines appear at stable position when Claude is working
- [x] Claude Code UI renders normally (no visual artifacts)
- [x] Lines disappear when Claude is idle
- [x] Claude Code output timing feels responsive
- [x] Unit tests cover cursor tracking and overlay generation

---

### Step 4: Static Emoji Placeholder ✓ COMPLETE

**Goal:** Render a stationary 🦖 on the game lines when Claude is working. Validates the overlay mechanism works with emoji rendering.

**Files modified:**
- `ui/model.go` - added `generateGameLines()` function and two-line overlay rendering

**Implementation (differs from original plan):**

1. Rendering logic is inline in `ui/model.go` via `generateGameLines(width)` function
2. Two-line layout:
   - Line 1 (sky): decorative clouds (`☁️`) for visual interest
   - Line 2 (ground): 🦖, obstacle placeholder (🌵), and score display
3. Example output:
   ```
   ☁️           ☁️                    ☁️                              [SKY]
   🦖                🌵                                           Score: 00000
   ```
4. Overlay positioned at `spinnerRow + 2` (below the CC status line)
5. Uses `RenderMultiLineOverlay()` from `inject/overlay.go`

**Key considerations:**
- `game/render.go` was NOT created - rendering kept in `ui/model.go` for now
- Will refactor to separate `game/render.go` in Step 5/6 when game state is introduced
- Minimum width handling: falls back to simple "☁️" / "🦖" if width < 20

**Acceptance criteria:**
- [x] 🦖 appears on game line when Claude is working
- [x] Emoji renders correctly (no visual glitches or misalignment)
- [x] Score placeholder visible
- [x] Ground line is visually distinct (sky + ground separation)
- [x] Two lines rendered (sky + ground)

**Status:** Complete - static placeholder working with two-line overlay

---

### Step 5: Game Core Logic ✓ COMPLETE

**Goal:** Implement game physics, collision detection, and scoring as pure functions with unit tests. No terminal/bubbletea code in this step.

**Files created:**
- `game/game.go` - game state struct, constants, and all core logic (physics, obstacles, collision, scoring)
- `game/render.go` - convert game state to two display lines (sky + ground)
- `game/game_test.go` - 23 unit tests covering all game logic and rendering

**Note:** Unlike the original plan which split logic across 5 files, all game logic lives in `game/game.go` and rendering in `game/render.go`. This is simpler and sufficient.

**Game state structure (as implemented):**

```go
type State struct {
    IsInAir        bool
    JumpTimeLeft   float64
    Obstacles      []float64  // obstacle x-positions (simplified from Obstacle struct)
    Score          int
    HighScore      int
    GameOver       bool
    IsPaused       bool
    ElapsedTime    float64    // total game time (for difficulty scaling)
    TimeSinceSpawn float64    // time since last obstacle spawn
}
```

**Core functions (all pure, no side effects):**

1. `Tick(s State, dt float64, width int) State` - advance game by dt seconds (includes obstacle movement, spawning, collision, scoring)
2. `Jump(s State) State` - initiate jump if on ground (sets `IsInAir`, `JumpTimeLeft`)
3. `Restart(s State) State` - reset game state while preserving high score
4. `TogglePause(s State) State` - pause/unpause (disabled when game over)
5. `Render(s State, width int) (skyLine, groundLine string)` - convert state to two display lines
6. `FormatCountdown(skyLine string, countdown int, width int) string` - overlay countdown on sky line

**Physics model (differs from original plan):**
- Uses duration-based jumping (`JumpDuration = 0.4s`) instead of velocity + gravity
- Jump arc is simulated via time-based height calculation in render
- `BaseObstacleSpeed = 40.0`, increases by `SpeedIncreaseRate = 2.0` per 10s
- Spawn interval decreases from `2.0s` to `0.8s` as score increases
- Collision hitbox: dino occupies x-positions 0 to 2.0, only collides when on ground

**Acceptance criteria:**
- [x] Dino can jump and land with duration-based arc
- [x] Obstacles spawn and move left with increasing difficulty
- [x] Collision detection works correctly (ground only, hitbox-based)
- [x] Score increments over time
- [x] All functions are pure (same input → same output)
- [x] 23 unit tests covering all game logic and rendering

**Status:** Complete

---

### Step 6: Game Rendering Integration ✓ COMPLETE

**Goal:** Wire up the existing game core (`game/` package) to the bubbletea model and terminal output. Replace the static `generateGameLines()` placeholder with live game state.

**Files modified:**
- `ui/model.go` - added `game.State` field, replaced `generateGameLines()` with `game.Render()`, added game tick (~50ms / 20 FPS), handle `GameKeyMsg`
- `main.go` - intercepts Space/R/P keypresses when game is active via `atomic.Bool`, passes others through to PTY

**Implementation:**

1. Added `game.State` and `lastGameTick` to bubbletea model
2. Added `gameTickMsg` at 50ms intervals calling `game.Tick()` when game is active
3. Added `GameKeyMsg` type with `KeyJump`, `KeyRestart`, `KeyPause` variants
4. `main.go` reads stdin byte-by-byte, checks `gameActive` atomic bool, intercepts space/r/R/p/P, forwards others to PTY
5. Game state initialized via `game.NewState()` in `NewModel()`

**Acceptance criteria:**
- [x] Dino jumps when space is pressed
- [x] Obstacles scroll from right to left
- [x] Collision ends the game (dino becomes 💀)
- [x] Score increments during play
- [x] R restarts the game
- [x] P pauses/unpauses
- [x] Non-game keys still work in Claude Code

**Status:** Complete

---

### Step 7: Pause/Resume with Countdown ✓ COMPLETE

**Goal:** Pause game when Claude stops working, resume with 3-second countdown when Claude starts working again.

**Files modified:**
- `ui/model.go` - added `countdownRemaining float64` field, auto-pause/resume logic, countdown rendering

**Implementation:**

1. Added `countdownRemaining float64` to `Model` (UI concern, not game state)
2. On idle transition: set `gameState.IsPaused = true` (unless game over)
3. On working transition: if game is paused and not game over, set `countdownRemaining = 3.0`
4. Each `gameTickMsg` decrements countdown; when it hits 0, unpauses the game
5. `renderOverlay()` calls `game.FormatCountdown()` when countdown > 0
6. All `GameKeyMsg` blocked during countdown
7. First activation has no countdown (game starts unpaused via `NewState()`)
8. Game over on resume skips countdown (`!GameOver` check is false)
9. Idle during countdown: overlay cleared, game stays paused; next working restart countdown from 3.0

**Acceptance criteria:**
- [x] Game pauses when Claude stops working
- [x] Game state preserved across pause/resume cycles
- [x] 3-second countdown displays before resuming
- [x] Countdown cancels if Claude stops working mid-countdown
- [x] Keys blocked during countdown

**Status:** Complete

---

### Step 8: Highscores

**Goal:** Persist high scores to disk, display in game UI.

**Files to create/modify:**
- `game/highscore.go` - load/save highscore to disk
- `ui/model.go` - load highscore on startup, save on game over

**What already exists (from Step 5):**
- `game.State.HighScore` field - in-memory tracking
- `game.Render()` already displays `HI: %05d` in the ground line
- `game.Tick()` already updates `HighScore` when `Score > HighScore`

**Storage:**
- File: `~/.claudinosaur/highscore` (simple text file with single integer)
- Create directory if it doesn't exist
- Handle missing/corrupted file gracefully (default to 0)

**Behavior:**

1. On startup: load highscore from disk
2. During game: track if current score > highscore
3. On game over: if new highscore, save to disk immediately
4. Display: `HI: 00099` always visible, flashes or highlights when beaten

**Functions:**

```go
func LoadHighScore() (int, error)
func SaveHighScore(score int) error
```

**Key considerations:**
- File I/O is impure → keep isolated from game core
- Don't block game loop on disk writes (but OK for this simple case)
- Highscore persists across sessions and across Claude Code restarts

**Acceptance criteria:**
- [ ] Highscore loads on startup
- [ ] Highscore saves when beaten
- [ ] Highscore persists after closing and reopening
- [ ] Graceful handling of missing/corrupted file

**Status:** Not started

---

### Step 9: Terminal Resize Handling

**Goal:** Handle terminal resize events gracefully during gameplay.

**Files to modify:**
- `main.go` - SIGWINCH handling (already forwarding to Claude)
- `ui/model.go` - track terminal width, re-render on resize
- `game/render.go` - adapt game width to terminal width

**Behavior:**

1. Track current terminal width
2. On resize (SIGWINCH):
   - Update stored width
   - Forward signal to Claude Code PTY (already done in step 1)
   - Re-render game lines with new width
3. Game adapts:
   - Minimum width: 60 chars (game unplayable below this)
   - Maximum width: no limit, game expands
   - Obstacles and dino positions are relative, scale with width

**Key considerations:**
- Bubbletea has built-in window size messages (`tea.WindowSizeMsg`)
- Game width affects obstacle spawn distance and visible play area
- If terminal too narrow, show warning instead of game?

**Acceptance criteria:**
- [ ] Game re-renders correctly after terminal resize
- [ ] No visual glitches during resize
- [ ] Claude Code also handles resize correctly (passthrough SIGWINCH)

**Status:** Not started

---

### Step 10: Hooks Refactor (TBC)

**Goal:** Replace spinner detection with Claude Code hooks for more reliable state detection.

**Why:**
- Spinner detection is fragile (depends on Claude Code's internal rendering)
- Hooks are an official API, less likely to break
- More precise timing for state changes

**Approach (to be refined):**

1. Research Claude Code hooks API
2. Create hooks that signal "work started" / "work ended"
3. Hooks communicate to Claudinosaur via:
   - Named pipe (FIFO)
   - Unix socket
   - File-based signaling
   - Environment variable + polling
4. Replace spinner-based detector with hook-based detector
5. Keep spinner detection as fallback option (`--use-spinner` flag)

**Open questions:**
- What hooks are available in Claude Code?
- What's the simplest IPC mechanism for this use case?
- Should both detection methods be supported?

**Status:** Not started - research needed

