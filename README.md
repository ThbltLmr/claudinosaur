# Claudinosaur

<!-- TODO: Add intro -->

<!-- TODO: Add demo video/gif -->

## Installation

```bash
go build -o claudinosaur
```

## Usage

```bash
./claudinosaur                          # Wraps Claude Code
./claudinosaur --test-cmd ./script.sh   # Test mode with custom command
```

## Controls

| Key | Action |
|-----|--------|
| `Space` | Jump |
| `R` | Restart |
| `P` | Pause |

## How it works

Claudinosaur is a PTY wrapper that detects Claude Code's working state via spinner characters and overlays a dinosaur game above the prompt. The game auto-pauses when Claude finishes and resumes with a countdown when Claude starts working again.
