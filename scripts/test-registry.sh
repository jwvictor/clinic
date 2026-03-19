#!/usr/bin/env bash
#
# Test every tool in the clinic registry inside a clean Docker environment.
#
# Three checks per tool:
#   1. Install — `clinic add <tool>` exits 0
#   2. Version — the tool's version_command runs and version_pattern matches
#   3. Auth    — the tool's auth_check command is a real subcommand (not "unknown command")
#
# Usage:
#   docker run --rm clinic-test /build/scripts/test-registry.sh [tool1 tool2 ...]
#
# If no tools are specified, all tools in the registry are tested.

set -euo pipefail

REGISTRY_DIR="/build/registry"
PASS=0
FAIL=0
SKIP=0
FAILURES=()

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
RESET='\033[0m'

ok()   { printf "${GREEN}✓${RESET}"; }
fail() { printf "${RED}✗${RESET}"; }
skip() { printf "${YELLOW}—${RESET}"; }

# Get tool list from args or from registry index
if [ $# -gt 0 ]; then
    TOOLS=("$@")
else
    TOOLS=($(python3 -c "
import json
idx = json.load(open('$REGISTRY_DIR/index.json'))
print(' '.join(sorted(idx['tools'])))
"))
fi

echo ""
printf "${BOLD}%-15s %-12s %-12s %-12s${RESET}\n" "TOOL" "INSTALL" "VERSION" "AUTH-CMD"
printf "%-15s %-12s %-12s %-12s\n" "────" "───────" "───────" "────────"

for tool_name in "${TOOLS[@]}"; do
    tool_file="$REGISTRY_DIR/tools/${tool_name}.json"
    if [ ! -f "$tool_file" ]; then
        printf "%-15s " "$tool_name"
        printf "$(fail) not in registry\n"
        FAIL=$((FAIL + 1))
        FAILURES+=("$tool_name: not found in registry")
        continue
    fi

    # Parse tool JSON
    cmd=$(python3 -c "import json; t=json.load(open('$tool_file')); print(t['command'])")
    version_cmd=$(python3 -c "import json; t=json.load(open('$tool_file')); print(t.get('version_command',''))")
    version_pattern=$(python3 -c "import json; t=json.load(open('$tool_file')); print(t.get('version_pattern',''))")
    auth_check=$(python3 -c "import json; t=json.load(open('$tool_file')); print(t.get('auth',{}).get('auth_check',''))")

    printf "%-15s " "$tool_name"

    # --- 1. Install ---
    install_result=""
    if command -v "$cmd" &>/dev/null; then
        install_result="pre-installed"
        printf "$(ok) %-10s " "exists"
    else
        if clinic add "$tool_name" &>/tmp/clinic-add-${tool_name}.log; then
            if command -v "$cmd" &>/dev/null; then
                install_result="ok"
                printf "$(ok) %-10s " "ok"
            else
                install_result="fail"
                printf "$(fail) %-10s " "no binary"
                FAIL=$((FAIL + 1))
                FAILURES+=("$tool_name: installed but '$cmd' not in PATH")
            fi
        else
            install_result="fail"
            printf "$(fail) %-10s " "failed"
            FAIL=$((FAIL + 1))
            # Grab last meaningful line from log
            reason=$(tail -5 /tmp/clinic-add-${tool_name}.log | grep -i -m1 'error\|fail\|not found\|no available' || tail -1 /tmp/clinic-add-${tool_name}.log)
            FAILURES+=("$tool_name: install failed — $reason")
        fi
    fi

    # --- 2. Version ---
    if [ "$install_result" = "fail" ]; then
        printf "$(skip) %-10s " "skip"
        SKIP=$((SKIP + 1))
    elif [ -z "$version_cmd" ]; then
        printf "$(skip) %-10s " "no cmd"
        SKIP=$((SKIP + 1))
    else
        # Run the version command
        version_output=$(eval "$version_cmd" 2>&1 || true)
        if [ -n "$version_pattern" ]; then
            matched=$(echo "$version_output" | python3 -c "
import sys, re
text = sys.stdin.read()
m = re.search(r'''$version_pattern''', text)
print(m.group(1) if m else '')
" 2>/dev/null || true)
            if [ -n "$matched" ]; then
                printf "$(ok) %-10s " "$matched"
                PASS=$((PASS + 1))
            else
                printf "$(fail) %-10s " "no match"
                FAIL=$((FAIL + 1))
                FAILURES+=("$tool_name: version_pattern didn't match output: $(echo "$version_output" | head -1)")
            fi
        else
            # No pattern, just check it runs
            if [ -n "$version_output" ]; then
                printf "$(ok) %-10s " "ok"
                PASS=$((PASS + 1))
            else
                printf "$(fail) %-10s " "empty"
                FAIL=$((FAIL + 1))
                FAILURES+=("$tool_name: version command produced no output")
            fi
        fi
    fi

    # --- 3. Auth check ---
    if [ "$install_result" = "fail" ]; then
        printf "$(skip) %-10s" "skip"
        SKIP=$((SKIP + 1))
    elif [ -z "$auth_check" ]; then
        printf "$(skip) %-10s" "none"
        SKIP=$((SKIP + 1))
    else
        # Run auth_check — we don't care about exit code (tool isn't authed),
        # we just want to confirm the subcommand is recognized.
        auth_output=$(eval "$auth_check" 2>&1 || true)
        auth_lower=$(echo "$auth_output" | tr '[:upper:]' '[:lower:]')

        # Check for signs the subcommand itself is invalid
        if echo "$auth_lower" | grep -qE 'unknown command|unrecognized|invalid (sub)?command|not a .* command|no such command|did you mean'; then
            printf "$(fail) %-10s" "bad cmd"
            FAIL=$((FAIL + 1))
            FAILURES+=("$tool_name: auth_check command not recognized — $(echo "$auth_output" | head -1)")
        else
            printf "$(ok) %-10s" "ok"
            PASS=$((PASS + 1))
        fi
    fi

    echo ""
done

# --- Summary ---
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
total=${#TOOLS[@]}
printf "Tested ${BOLD}%d${RESET} tools: " "$total"
printf "${GREEN}%d passed${RESET}, " "$PASS"
printf "${RED}%d failed${RESET}, " "$FAIL"
printf "${YELLOW}%d skipped${RESET}\n" "$SKIP"

if [ ${#FAILURES[@]} -gt 0 ]; then
    echo ""
    printf "${RED}${BOLD}Failures:${RESET}\n"
    for f in "${FAILURES[@]}"; do
        printf "  ${RED}•${RESET} %s\n" "$f"
    done
    echo ""
    exit 1
fi

echo ""
exit 0
