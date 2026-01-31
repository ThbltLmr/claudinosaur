#!/bin/bash
# Capture screenshots of the game running in tmux
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
TIMEOUT=60
INTERVAL=2

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
    tmux send-keys -t "$SESSION_NAME" "write a detailed 500 word essay about the history of programming languages from the 1950s to today"
    sleep 0.5
    tmux send-keys -t "$SESSION_NAME" Escape
    sleep 0.3
    tmux send-keys -t "$SESSION_NAME" Enter

    echo "Waiting for Claude to start working..."
    sleep 2
else
    echo "Starting claudinosaur with mock claude..."

    # Run claudinosaur with mock claude
    tmux send-keys -t "$SESSION_NAME" "$PROJECT_DIR/claudinosaur --test-cmd $PROJECT_DIR/scripts/mock_claude.sh" Enter

    # Wait for the game to render
    sleep 3
fi

# Create a session-specific directory for this capture run
RUN_ID=$(date +%Y%m%d_%H%M%S)
RUN_DIR="$SCREENSHOT_DIR/run_${MODE#--}_${RUN_ID}"
mkdir -p "$RUN_DIR"

echo ""
echo "Capturing screenshots to: $RUN_DIR"
echo "Timeout: ${TIMEOUT}s, Interval: ${INTERVAL}s"
echo ""

# Capture screenshots over time
START_TIME=$(date +%s)
SCREENSHOT_COUNT=0
GAME_DETECTED=false
GAME_ENDED=false

while true; do
    CURRENT_TIME=$(date +%s)
    ELAPSED=$((CURRENT_TIME - START_TIME))

    # Check timeout
    if [ $ELAPSED -ge $TIMEOUT ]; then
        echo "Timeout reached (${TIMEOUT}s)"
        break
    fi

    # Capture current pane content
    CONTENT=$(tmux capture-pane -t "$SESSION_NAME" -p 2>/dev/null || echo "SESSION_ENDED")

    if [ "$CONTENT" = "SESSION_ENDED" ]; then
        echo "Session ended"
        break
    fi

    # Check if game is visible
    if echo "$CONTENT" | grep -q "🦕\|🦖\|🌵"; then
        if [ "$GAME_DETECTED" = false ]; then
            echo "Game started!"
            GAME_DETECTED=true
        fi

        # Save screenshot
        SCREENSHOT_COUNT=$((SCREENSHOT_COUNT + 1))
        SCREENSHOT_FILE="$RUN_DIR/frame_$(printf '%03d' $SCREENSHOT_COUNT).txt"
        echo "$CONTENT" > "$SCREENSHOT_FILE"
        echo "[$ELAPSED s] Captured frame $SCREENSHOT_COUNT"
    elif [ "$GAME_DETECTED" = true ]; then
        # Game was visible but now it's not - Claude finished
        echo "Game ended (Claude finished working)"
        GAME_ENDED=true

        # Capture final state
        SCREENSHOT_COUNT=$((SCREENSHOT_COUNT + 1))
        SCREENSHOT_FILE="$RUN_DIR/frame_$(printf '%03d' $SCREENSHOT_COUNT)_final.txt"
        echo "$CONTENT" > "$SCREENSHOT_FILE"
        echo "[$ELAPSED s] Captured final frame $SCREENSHOT_COUNT"
        break
    fi

    sleep $INTERVAL
done

# Clean up
tmux kill-session -t "$SESSION_NAME" 2>/dev/null || true

echo ""
echo "=========================================="
echo "Capture complete!"
echo "Screenshots: $SCREENSHOT_COUNT"
echo "Directory: $RUN_DIR"
echo "=========================================="

# Show first and last frames if we have any
if [ $SCREENSHOT_COUNT -gt 0 ]; then
    FIRST_FRAME=$(ls "$RUN_DIR"/frame_*.txt 2>/dev/null | head -1)
    LAST_FRAME=$(ls "$RUN_DIR"/frame_*.txt 2>/dev/null | tail -1)

    echo ""
    echo "First frame:"
    echo "----------------------------------------"
    cat "$FIRST_FRAME"
    echo "----------------------------------------"

    if [ "$FIRST_FRAME" != "$LAST_FRAME" ]; then
        echo ""
        echo "Last frame:"
        echo "----------------------------------------"
        cat "$LAST_FRAME"
        echo "----------------------------------------"
    fi
fi
