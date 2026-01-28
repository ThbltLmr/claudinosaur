#!/bin/bash
# Capture a screenshot of the game running in tmux
#
# Usage:
#   ./capture_screenshot.sh          # Uses mock claude (fast, no API calls)
#   ./capture_screenshot.sh --real   # Uses real Claude Code (requires API key)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
SCREENSHOT_DIR="$PROJECT_DIR/.claude/screenshots"
SESSION_NAME="claudinosaur-test"
MODE="${1:-mock}"

mkdir -p "$SCREENSHOT_DIR"

# Kill any existing session
tmux kill-session -t "$SESSION_NAME" 2>/dev/null || true

# Start a new tmux session in detached mode
tmux new-session -d -s "$SESSION_NAME" -x 120 -y 30

if [ "$MODE" = "--real" ]; then
    echo "Starting claudinosaur with real Claude Code..."

    # Run claudinosaur (wraps real claude)
    tmux send-keys -t "$SESSION_NAME" "$PROJECT_DIR/claudinosaur" Enter

    # Wait for Claude Code to fully initialize
    echo "Waiting for Claude Code to initialize..."
    sleep 10

    # Enter INSERT mode, type prompt, exit INSERT mode, submit
    echo "Sending prompt..."
    tmux send-keys -t "$SESSION_NAME" "i"
    sleep 0.2
    tmux send-keys -t "$SESSION_NAME" "write a detailed 500 word essay about the history of programming languages"
    sleep 0.5
    tmux send-keys -t "$SESSION_NAME" Escape
    sleep 0.3
    tmux send-keys -t "$SESSION_NAME" Enter

    # Wait for Claude to start working (spinner appears) and game to render
    echo "Waiting for Claude to start working..."
    sleep 2

    # Poll for game content (dinosaur emoji)
    for i in {1..15}; do
        CONTENT=$(tmux capture-pane -t "$SESSION_NAME" -p)
        if echo "$CONTENT" | grep -q "🦖\|🌵"; then
            echo "Game detected!"
            break
        fi
        echo "Waiting for game... ($i/15)"
        sleep 1
    done
else
    echo "Starting claudinosaur with mock claude..."

    # Run claudinosaur with mock claude
    tmux send-keys -t "$SESSION_NAME" "$PROJECT_DIR/claudinosaur --test-cmd $PROJECT_DIR/scripts/mock_claude.sh" Enter

    # Wait for the game to render
    sleep 3
fi

# Capture the pane content
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
SCREENSHOT_FILE="$SCREENSHOT_DIR/screenshot_${MODE#--}_${TIMESTAMP}.txt"

tmux capture-pane -t "$SESSION_NAME" -p > "$SCREENSHOT_FILE"

echo "Screenshot saved to: $SCREENSHOT_FILE"
echo ""
echo "Content:"
echo "----------------------------------------"
cat "$SCREENSHOT_FILE"
echo "----------------------------------------"

# Clean up
tmux kill-session -t "$SESSION_NAME" 2>/dev/null || true

echo ""
echo "Done!"
