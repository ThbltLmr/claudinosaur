# Claudinosaur

For when you're sitting in an open space so you have to keep staring at your terminal while Claude is doing your job

https://github.com/user-attachments/assets/c01c6850-a82c-4c41-9a94-5557227902b0


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
