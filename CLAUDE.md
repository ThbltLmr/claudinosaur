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

         🌵💀                                                                       Score: 00060 HI: 00060 [R]estart  // DINO GAME HERE!!
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
- game active (replace the two lines above the prompt input with the game content)

Trigger to switch between both to be decided:
- option 1: detect the "spinner" characters used by Claude Code
- option 2: use Claude Code hooks to communicate to the PTY (more setup but more reliable)

Shutdown:
- the application should be stopped with <ctrl-C>. It is okay to overwrite the Claude Code <ctrl-C> behaviour, since users can alternatively use <Esc> to interrupt Claude Code actions anyway.
- pressing <ctrl-C> in Claudinosaur should shutdown the underlying Claude Code instance, shut down the wrapper, and return to default terminal behaviour.

## Architecture

Follow a "functional core, imperative shell" approach. Code related to physics, scoring, rendering of the two lines of the game should be isolated from the bubbletea specific logic.
The core part should be pure functions covered by unit tests.

## Integration testing - To be investigated

Open question: have a "--test" mode where the whole rendered screen is compared to snapshots instead of being passed to Claude Code to prevent regressions?

## Coding style

Comments should be exceptional. Only write comments when it is unclear WHY a piece of code was needed. If you need a comment to explain what your code does, then your code is not explicit enough and that is a problem.

# Gameplay

## Controls

- <space> to jump
- <R> to restart
- <P> to pause

## Automatic pausing

When Claude Code is done thinking, the game should be paused, and the app should store current state. When Claude Code resume working, the game should re-appear and there should be a 3 second countdown before it restarts
