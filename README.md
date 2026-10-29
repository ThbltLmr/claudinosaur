# Claudinosaur

For when you're sitting in an open space so you have to keep staring at your terminal while Claude is doing your job

https://github.com/user-attachments/assets/259f5d3d-5828-4d5b-98a6-2ab2590bca8e


## Installation

Claudinosaur is a Claude Code mod. At the prompt of a Claude Code session:

```
/plugin install claudinosaur --marketplace ThbltLmr/claudinosaur
```

Answer `y` to add the marketplace, then pick a scope. From a shell:

```bash
claude plugin install claudinosaur --marketplace ThbltLmr/claudinosaur
```

To run it from a clone instead, without installing:

```bash
claude --plugin-dir .
```

## Controls

While Claude works, press `ctrl+x tab` to focus the game. `Esc` hands the keyboard back to the prompt, which keeps working normally while the game is unfocused.

| Key | Action |
|-----|--------|
| `Space` | Jump |
| `R` | Restart |
| `P` | Pause |

## How it works

Claude Code draws the game in the band above the prompt while a turn runs. The game pauses when Claude finishes and resumes with a countdown when Claude starts working again. The high score is kept across sessions.

## Development

```bash
claude plugin validate .                # What Claude Code sees and would refuse
claude plugin test .                    # Unit and band tests
./scripts/capture_mod.sh                # Frames of a real Claude Code session with the mod
```
