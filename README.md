# Claudinosaur

For when you're sitting in an open space so you have to keep staring at your terminal while Claude is doing your job

https://github.com/user-attachments/assets/259f5d3d-5828-4d5b-98a6-2ab2590bca8e


## As a Claude Code mod (experimental)

The `mod/` folder is the same game as a Claude Code mod: no wrapper, Claude Code draws it in the band above the prompt while a turn runs.

```bash
claude --plugin-dir ./mod
```

While Claude works, press `ctrl+x tab` to focus the game, then `Space` to jump, `P` to pause and `R` to restart. `Esc` hands the keyboard back to the prompt, which keeps working normally while the game is unfocused. The high score is kept across sessions.

```bash
claude plugin test ./mod                # Unit and band tests
./scripts/capture_mod.sh                # Frames of a real Claude Code session with the mod
```

## Installation

```bash
go build -o claudinosaur
```

## Usage

```bash
./claudinosaur                          # Wraps Claude Code
```

## Controls

| Key | Action |
|-----|--------|
| `Space` | Jump |
| `R` | Restart |
| `P` | Pause |

## How it works

Claudinosaur is a PTY wrapper that detects Claude Code's working state via spinner characters and overlays a dinosaur game above the prompt. The game auto-pauses when Claude finishes and resumes with a countdown when Claude starts working again.
