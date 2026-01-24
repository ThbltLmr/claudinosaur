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

### Step 2: State Detection ✧ CURRENT

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
4. Log state transitions to stderr or a debug file (e.g., `[STATE] idle → working`)
5. Add a `-debug` flag to enable/disable state logging

**Key considerations:**
- Output arrives in chunks, but spinner chars are single Unicode codepoints → simpler than pattern matching
- Spinner disappears when Claude is done → absence of spinner for N milliseconds = idle
- Must not slow down or corrupt the passthrough (detector is read-only observer)

**Acceptance criteria:**
- [ ] State changes are logged when Claude starts/stops working
- [ ] Detection works across different Claude actions (thinking, tool use, subagents)
- [ ] PTY passthrough remains fully functional
- [ ] Unit tests cover spinner character detection and timeout logic

**Open question:**
- What's a good timeout for "spinner gone = idle"? (Start with 500ms, tune later)

---

### Step 3: Terminal Injection Mechanism

**Goal:** Inject two lines above the prompt separator when Claude is working. Test with static placeholder text.

**Files to modify/create:**
- `main.go` - integrate bubbletea, manage rendering mode
- `ui/model.go` - bubbletea model with passthrough + game modes
- `ui/injector.go` - output stream transformation logic
- `ui/injector_test.go` - unit tests

**Injection point detection:**

Parse output stream looking for this sequence:
1. Spinner character (one of: `✢✶✻✸✹✺✷`)
2. Rest of line (status text like "Zigzagging... (Esc to interrupt)")
3. Newline
4. Line of dashes (`─────...`) = prompt separator

Inject our two game lines BEFORE the dash line.

**Technical approach:**

1. Create an `Injector` that buffers output and transforms it:
   - Pass through everything normally until we detect spinner + approaching dash line
   - When we see the dash line coming, insert game lines first, then the dashes
2. Two modes:
   - `passthrough`: no transformation
   - `gameActive`: inject game lines before each dash separator
3. Bubbletea manages the mode switching based on state detector events
4. Test with static text: "=== GAME LINE 1 ===" and "=== GAME LINE 2 ==="

**Key considerations:**
- Need to buffer enough output to detect the pattern (spinner → newline → dashes)
- Dashes line might arrive in a separate chunk from the spinner → buffering required
- Must flush buffer promptly to avoid laggy output
- The injector should be a pure function: `(buffer, newChunk, mode) → (outputToWrite, newBuffer)`

**Acceptance criteria:**
- [ ] Two placeholder lines appear above the prompt separator when Claude is working
- [ ] Lines are not injected when Claude is idle
- [ ] Claude Code output timing feels responsive (no noticeable lag)
- [ ] Works correctly when output arrives in various chunk sizes
- [ ] Unit tests cover pattern detection and injection logic

**Status:** Not started

---

### Step 4: Static Emoji Placeholder

**Goal:** Render a stationary 🦖 on the game lines when Claude is working. Validates the injection mechanism works with emoji rendering.

**Files to modify/create:**
- `ui/injector.go` - update placeholder text to emoji layout
- `game/render.go` - game line rendering (initially just static dino)

**Technical approach:**

1. Replace static placeholder text with actual game line format:
   - Line 1: empty or "air" row (where dino appears when jumping)
   - Line 2: ground row with 🦖 and ground characters
2. Basic layout (example, ~80 chars wide):
   ```
   [                                                                    ]  <- air row
   [🦖__________________________________________________________________]  <- ground row
   ```
3. Add score display area on the right side of ground row:
   ```
   🦖________________________________________________________________ Score: 00000
   ```

**Key considerations:**
- Emoji width: 🦖 may render as 2 columns in terminal → need to account for this
- Terminal width: game width should adapt or have a minimum width
- Keep rendering logic separate from injection logic

**Acceptance criteria:**
- [ ] 🦖 appears on game line when Claude is working
- [ ] Emoji renders correctly (no visual glitches or misalignment)
- [ ] Score placeholder visible
- [ ] Ground line is visually distinct

**Status:** Not started

---

### Step 5: Game Core Logic

**Goal:** Implement game physics, collision detection, and scoring as pure functions with unit tests. No terminal/bubbletea code in this step.

**Files to create:**
- `game/state.go` - game state struct and constants
- `game/physics.go` - jump mechanics, gravity, position updates
- `game/obstacles.go` - obstacle spawning and movement
- `game/collision.go` - collision detection
- `game/score.go` - scoring logic
- `game/*_test.go` - unit tests for each module

**Game state structure:**

```go
type GameState struct {
    DinoY        float64    // 0 = ground, positive = air
    DinoVelocity float64    // vertical velocity
    IsJumping    bool
    Obstacles    []Obstacle // positions of cacti
    Score        int
    GameOver     bool
    IsPaused     bool
}

type Obstacle struct {
    X     float64 // horizontal position, decreases over time
    Width int     // collision width
}
```

**Core functions (all pure, no side effects):**

1. `Tick(state GameState, dt float64) GameState` - advance game by dt seconds
2. `Jump(state GameState) GameState` - initiate jump if on ground
3. `SpawnObstacle(state GameState) GameState` - add new obstacle at right edge
4. `CheckCollision(state GameState) bool` - dino hit an obstacle?
5. `UpdateScore(state GameState) GameState` - increment score based on distance

**Physics parameters:**
- Gravity: tune for game feel (start with ~1500 units/s²)
- Jump velocity: tune for jump height (start with ~500 units/s)
- Obstacle speed: increases over time for difficulty scaling
- Ground level: Y = 0

**Acceptance criteria:**
- [ ] Dino can jump and land with gravity
- [ ] Obstacles spawn and move left
- [ ] Collision detection works correctly
- [ ] Score increments over time
- [ ] All functions are pure (same input → same output)
- [ ] >90% test coverage on game logic

**Status:** Not started

---

### Step 6: Game Rendering Integration

**Goal:** Connect game core to terminal output. Full playable game with obstacles, jumping, and score display.

**Files to modify/create:**
- `game/render.go` - convert GameState to two display lines
- `game/render_test.go` - snapshot tests for rendering
- `ui/model.go` - wire up game tick, input handling
- `main.go` - handle game input (space, R, P)

**Rendering function:**

```go
func Render(state GameState, width int) (line1, line2 string)
```

Converts game state to two strings:
- Line 1 (air): shows 🦖 when jumping high enough
- Line 2 (ground): shows 🦖 when on/near ground, obstacles (🌵), score

**Visual elements:**
- Dino: 🦖 (or 💀 when game over)
- Cactus: 🌵
- Ground: `_` or `▁` characters
- Empty space: spaces

**Example renders:**
```
Playing (on ground):
                                                                Score: 00042
🦖_______________________________🌵______________🌵_____________ HI: 00099

Playing (jumping):
        🦖                                                      Score: 00042
________________________________🌵______________🌵_____________ HI: 00099

Game over:
                                                                Score: 00042
_______________________________🌵💀______________🌵_____________ HI: 00099 [R]estart
```

**Input handling:**
- Space: call `Jump()` on game state
- R: reset game state
- P: toggle pause
- All other keys: pass through to Claude Code

**Game loop:**
- Bubbletea tick every ~50ms (20 FPS) when game is active
- Each tick: `state = Tick(state, 0.05)`
- After tick: re-render and update injected lines

**Acceptance criteria:**
- [ ] Dino jumps when space is pressed
- [ ] Obstacles scroll from right to left
- [ ] Collision ends the game (dino becomes 💀)
- [ ] Score increments during play
- [ ] R restarts the game
- [ ] P pauses/unpauses
- [ ] Non-game keys still work in Claude Code

**Status:** Not started

---

### Step 7: Pause/Resume with Countdown

**Goal:** Pause game when Claude stops working, resume with 3-second countdown when Claude starts working again.

**Files to modify:**
- `game/state.go` - add countdown state
- `game/render.go` - render countdown overlay
- `ui/model.go` - handle state transitions with countdown logic

**State additions:**

```go
type GameState struct {
    // ... existing fields
    CountdownSeconds int  // 3, 2, 1, 0 (0 = playing)
}
```

**Behavior:**

1. Claude stops working (idle):
   - Game pauses immediately
   - Game lines disappear (back to passthrough mode)
   - Game state is preserved in memory

2. Claude starts working again:
   - Game lines reappear
   - Show countdown: "3..." → "2..." → "1..." → resume
   - Countdown ticks every 1 second
   - Game resumes after countdown reaches 0

3. If Claude stops working during countdown:
   - Cancel countdown
   - Hide game lines
   - Next time Claude works, restart countdown from 3

**Countdown rendering:**
```
                           ▶ 3 ◀                                Score: 00042
🦖_______________________________🌵______________🌵_____________ HI: 00099
```

**Key considerations:**
- Countdown uses separate 1-second tick (not game tick)
- Game tick only runs when CountdownSeconds == 0
- Obstacles and dino position frozen during pause/countdown

**Acceptance criteria:**
- [ ] Game pauses when Claude stops working
- [ ] Game state preserved across pause/resume cycles
- [ ] 3-second countdown displays before resuming
- [ ] Countdown cancels if Claude stops working mid-countdown
- [ ] Player can still press R to restart during countdown

**Status:** Not started

---

### Step 8: Highscores

**Goal:** Persist high scores to disk, display in game UI.

**Files to create/modify:**
- `game/highscore.go` - load/save highscore
- `game/state.go` - add HighScore field
- `game/render.go` - already displays HI: score

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

