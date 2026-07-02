#!/usr/bin/env bash
# Overload: hammer the system with RATE rps (default 5000, ~5x steady state)
# for DURATION (default 2m). Which component gives out first?
# shellcheck source=chaos-toolkit/scripts/_lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

RATE="${RATE:-5000}"
DURATION="${DURATION:-2m}"
banner "overload — ${RATE} rps for ${DURATION}"

echo "While it runs, keep these panels open:"
echo "  - Golden Signals: which signal degrades FIRST?"
echo "  - Saturation panels: which resource is the ceiling?"
echo

dc run --rm -e RATE="$RATE" -e DURATION="$DURATION" k6 run /scripts/steady-state.js

echo
echo "Post-mortem prompts (journal them):"
echo "  - What was the breaking point in rps?"
echo "  - What would you scale first, and what evidence says so?"
