#!/bin/bash
# Mock claude that outputs spinner characters to trigger the game overlay

SPINNERS="✢✶✻✸✹✺✷"

echo "Mock Claude - outputting spinners to trigger game..."
echo ""

i=0
while true; do
    char="${SPINNERS:$((i % 7)):1}"
    # Move to line 5, clear line, print spinner
    echo -ne "\033[5;1H\033[2K* Working... $char"
    i=$((i + 1))
    sleep 0.1
done
