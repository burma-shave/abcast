#!/usr/bin/env bash
# publish.sh — convert an m4b audiobook and publish it to Azure Blob Storage
# Usage: ./publish.sh <storage-account> <audiobook.m4b>

set -euo pipefail

BOLD="\033[1m"
CYAN="\033[36m"
RED="\033[31m"
ORANGE="\033[38;5;208m"
DIM="\033[2m"
RESET="\033[0m"

fatal() {
    echo -e "\n${BOLD}${RED}✗  FELOTA — $*${RESET}"
    echo -e "${RED}\"Fix it or don't come back.\"${RESET}\n"
    exit 1
}

if [[ $# -lt 2 ]]; then
    echo -e "\n${RED}\"Oye. Account and file. Two things. Sa sa ke?\"${RESET}"
    echo -e "\nUsage: $0 <storage-account> <audiobook.m4b>\n"
    exit 1
fi

ACCOUNT="$1"
M4B="$2"

[[ -f "$M4B" ]] || fatal "file not found: $M4B"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Derive container name from the audiobook filename:
# lowercase, spaces/underscores to hyphens, strip non-alphanumeric-hyphen chars.
STEM="$(basename "$M4B" .m4b)"
CONTAINER="$(echo "$STEM" | tr '[:upper:]' '[:lower:]' | tr ' _' '-' | tr -cd 'a-z0-9-' | sed 's/^-*//' | cut -c1-63)"

[[ -n "$CONTAINER" ]] || fatal "could not derive a valid container name from filename: $STEM"

BASE_URL="https://${ACCOUNT}.blob.core.windows.net/${CONTAINER}"
OUT_DIR="$(dirname "$M4B")/${STEM}"

width=58
line=$(printf '─%.0s' $(seq 1 $width))
echo -e "\n${BOLD}${ORANGE}${line}${RESET}"
echo -e "  ${BOLD}ABCAST — Publish${RESET}"
echo -e "${BOLD}${ORANGE}${line}${RESET}\n"
echo -e "   ${DIM}\"Two commands. One mission. No felota.\"\n${RESET}"
echo -e "   ${BOLD}account  :${RESET} $ACCOUNT"
echo -e "   ${BOLD}container:${RESET} $CONTAINER"
echo -e "   ${BOLD}base URL :${RESET} $BASE_URL"
echo -e "   ${BOLD}output   :${RESET} $OUT_DIR\n"

# ── Step 1: convert ───────────────────────────────────────────────────────────

"${SCRIPT_DIR}/abcast" \
    -file  "$M4B" \
    -url   "$BASE_URL" \
    -out   "$OUT_DIR"

# ── Step 2: upload ────────────────────────────────────────────────────────────

"${SCRIPT_DIR}/upload.sh" \
    -a "$ACCOUNT" \
    -c "$CONTAINER" \
    -d "$OUT_DIR"

# ── Done ──────────────────────────────────────────────────────────────────────

echo -e "   ${DIM}\"Subscribe or don't. The feed is there either way.\"${RESET}"
echo -e "\n   ${BOLD}${CYAN}Feed URL:${RESET}  ${BASE_URL}/feed.xml\n"
