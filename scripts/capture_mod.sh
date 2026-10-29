#!/bin/bash
# Drives real Claude Code with the claudinosaur mod loaded through a full
# lifecycle (idle, working, jump, pause, idle, resume) and saves one frame per step.
#
# Usage: ./scripts/capture_mod.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
SESSION="claudinosaur-mod"
RUN_DIR="$PROJECT_DIR/.claude/screenshots/mod_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$RUN_DIR"

capture() {
    tmux capture-pane -p -t "$SESSION" > "$RUN_DIR/$1.txt"
    echo "captured $1"
}

band_visible() {
    tmux capture-pane -p -t "$SESSION" | grep -q "␣ jump"
}

wait_until() {
    local condition=$1 limit=$2
    for _ in $(seq 1 "$limit"); do
        if eval "$condition"; then
            return 0
        fi
        sleep 0.5
    done
    echo "timed out waiting for: $condition" >&2
    return 1
}

submit() {
    tmux send-keys -t "$SESSION" -l "$1"
    sleep 0.3
    tmux send-keys -t "$SESSION" Enter
}

tmux kill-session -t "$SESSION" 2>/dev/null || true
tmux new-session -d -s "$SESSION" -x 120 -y 32 -c "$PROJECT_DIR" \
    "claude --plugin-dir $PROJECT_DIR --allowedTools Read"
trap 'tmux kill-session -t "$SESSION" 2>/dev/null || true' EXIT

sleep 8
capture 01_idle_no_band

submit "Read game/game.go, game/render.go, ui/model.go and main.go one at a time, then write a 300-word explanation of how this project works."
wait_until band_visible 40
sleep 2
capture 02_working_band_shown

tmux send-keys -t "$SESSION" C-x Tab
sleep 0.3
tmux send-keys -t "$SESSION" ' '
sleep 0.1
capture 03_space_jumps

sleep 1
tmux send-keys -t "$SESSION" p
sleep 0.5
capture 04_paused
sleep 1.5
capture 05_still_paused

tmux send-keys -t "$SESSION" p
sleep 1
capture 06_unpaused

tmux send-keys -t "$SESSION" Escape
wait_until '! band_visible' 240
capture 07_idle_band_hidden

submit "Read README.md, then reply with a 150-word summary of it."
wait_until band_visible 40
capture 08_resume_countdown
sleep 4
capture 09_resumed_after_countdown

wait_until '! band_visible' 240
capture 10_idle_again

echo "frames in $RUN_DIR"
