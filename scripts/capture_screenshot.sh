#!/bin/bash
# Capture a screenshot of the game running in tmux

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
SCREENSHOT_DIR="$PROJECT_DIR/.claude/screenshots"
SESSION_NAME="claudinosaur-test"

mkdir -p "$SCREENSHOT_DIR"

# Kill any existing session
tmux kill-session -t "$SESSION_NAME" 2>/dev/null || true

# Start a new tmux session in detached mode
tmux new-session -d -s "$SESSION_NAME" -x 120 -y 30

# Run claudinosaur with mock claude
tmux send-keys -t "$SESSION_NAME" "$PROJECT_DIR/claudinosaur --test-cmd $PROJECT_DIR/scripts/mock_claude.sh" Enter

# Wait for the game to render (spinner needs to be detected, game needs to start)
sleep 3

# Capture the pane content
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
SCREENSHOT_FILE="$SCREENSHOT_DIR/screenshot_${TIMESTAMP}.txt"

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
