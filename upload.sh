#!/usr/bin/env bash
# upload.sh — push abcast output to Azure Blob Storage
#
# Regular container:  ./upload.sh -a <account> -c <container> -d <output-dir>
# Static website:     ./upload.sh -a <account> -c '$web'      -d <output-dir>

set -euo pipefail

# ── helpers ───────────────────────────────────────────────────────────────────

BOLD="\033[1m"
CYAN="\033[36m"
GREEN="\033[32m"
YELLOW="\033[33m"
RED="\033[31m"
ORANGE="\033[38;5;208m"
DIM="\033[2m"
RESET="\033[0m"

step()  { echo -e "\n${BOLD}${CYAN}▶  $*${RESET}"; }
info()  { echo -e "   ${BOLD}$*${RESET}"; }
detail(){ echo -e "   ${DIM}$*${RESET}"; }
done_(){ echo -e "   ${BOLD}${GREEN}✓${RESET}  $*"; }
fatal() {
    echo -e "\n${BOLD}${RED}✗  FELOTA — $*${RESET}"
    echo -e "${RED}\"Fix it or don't come back.\"${RESET}\n"
    exit 1
}

drummer_start() {
    local width=58
    local line
    line=$(printf '─%.0s' $(seq 1 $width))
    echo -e "\n${BOLD}${ORANGE}${line}${RESET}"
    echo -e "  ${BOLD}ABCAST — Azure Upload${RESET}"
    echo -e "${BOLD}${ORANGE}${line}${RESET}\n"
    echo -e "   ${DIM}\"Milowda send it to the cloud. The void is just storage.\"\n${RESET}"
}

drummer_done() {
    local url="$1"
    echo -e "\n   ${DIM}\"It's up there. Floating in the black. Sa sa ke?\"${RESET}"
    echo -e "   Feed URL: ${BOLD}${CYAN}${url}/feed.xml${RESET}\n"
}

usage() {
    echo "Usage: $0 -a <storage-account> -c <container> -d <output-dir> [-s <subscription>]"
    echo
    echo "  -a  Azure storage account name           (required)"
    echo "  -c  Blob container name                  (required)"
    echo "      Use '\$web' to enable static website hosting"
    echo "  -d  abcast output directory              (required)"
    echo "  -s  Azure subscription name/ID           (optional)"
    exit 1
}

# ── args ──────────────────────────────────────────────────────────────────────

ACCOUNT=""
CONTAINER=""
OUT_DIR=""
SUBSCRIPTION=""

while getopts "a:c:d:s:" opt; do
    case $opt in
        a) ACCOUNT="$OPTARG" ;;
        c) CONTAINER="$OPTARG" ;;
        d) OUT_DIR="$OPTARG" ;;
        s) SUBSCRIPTION="$OPTARG" ;;
        *) usage ;;
    esac
done

[[ -z "$ACCOUNT"   ]] && { echo -e "\n${RED}\"No storage account. Deting.\"\n${RESET}"; usage; }
[[ -z "$CONTAINER" ]] && { echo -e "\n${RED}\"No container. Deting.\"\n${RESET}"; usage; }
[[ -z "$OUT_DIR"   ]] && { echo -e "\n${RED}\"No output directory. Deting.\"\n${RESET}"; usage; }
[[ -d "$OUT_DIR"   ]] || fatal "output directory not found: $OUT_DIR"

STATIC_SITE=false
[[ "$CONTAINER" == '$web' ]] && STATIC_SITE=true

drummer_start

# ── preflight ─────────────────────────────────────────────────────────────────

step "Preflight checks"

command -v az &>/dev/null || fatal "az CLI not found — install it first: https://aka.ms/install-azure-cli"
detail "az CLI found"

if [[ -n "$SUBSCRIPTION" ]]; then
    az account set --subscription "$SUBSCRIPTION" \
        || fatal "could not set subscription: $SUBSCRIPTION"
    detail "subscription set: $SUBSCRIPTION"
fi

az account show &>/dev/null \
    || fatal "not logged in — run: az login"
detail "logged in as: $(az account show --query "user.name" -o tsv)"
detail "subscription : $(az account show --query "name" -o tsv)"

done_ "preflight OK"

# ── container / static site setup ────────────────────────────────────────────

if [[ "$STATIC_SITE" == true ]]; then
    step "Enabling static website hosting"
    info "account   : $ACCOUNT"
    info "container : \$web"

    az storage blob service-properties update \
        --account-name "$ACCOUNT" \
        --static-website \
        --index-document "index.html" \
        --auth-mode login \
        --output none \
        || fatal "could not enable static website hosting"

    # $web is created automatically; fetch the web endpoint.
    ENDPOINT=$(az storage account show \
        --name "$ACCOUNT" \
        --query "primaryEndpoints.web" \
        -o tsv | tr -d '/')

    done_ "static website enabled"
    detail "endpoint: $ENDPOINT"
else
    step "Ensuring container exists"
    info "account   : $ACCOUNT"
    info "container : $CONTAINER"

    EXISTS=$(az storage container exists \
        --account-name "$ACCOUNT" \
        --name "$CONTAINER" \
        --auth-mode login \
        --query "exists" -o tsv 2>/dev/null)

    if [[ "$EXISTS" == "true" ]]; then
        detail "container already exists"
    else
        detail "creating container..."
        az storage container create \
            --account-name "$ACCOUNT" \
            --name "$CONTAINER" \
            --public-access blob \
            --auth-mode login \
            --output none \
            || fatal "could not create container"
        done_ "container created"
    fi

    az storage container set-permission \
        --account-name "$ACCOUNT" \
        --name "$CONTAINER" \
        --public-access blob \
        --auth-mode login \
        --output none \
        || fatal "could not set container permissions"

    ENDPOINT="https://${ACCOUNT}.blob.core.windows.net/${CONTAINER}"
    done_ "container ready"
fi

# ── upload ────────────────────────────────────────────────────────────────────

step "Uploading files"
echo -e "\n   ${DIM}\"Every file goes up. No exceptions. Beltalowda don't leave cargo behind.\"\n${RESET}"

upload_batch() {
    local pattern="$1"
    local content_type="$2"
    local label="$3"

    info "$label"
    az storage blob upload-batch \
        --account-name "$ACCOUNT" \
        --destination "$CONTAINER" \
        --source "$OUT_DIR" \
        --pattern "$pattern" \
        --content-type "$content_type" \
        --overwrite true \
        --auth-mode login \
        --output none \
        || fatal "upload failed for $pattern"
    done_ "$label"
}

upload_batch "*.xml"        "application/rss+xml; charset=utf-8"  "feed.xml"
upload_batch "*.html"       "text/html; charset=utf-8"            "index.html"
upload_batch "audio/*.m4a"  "audio/mp4"                           "audio chapters"

# ── verify & report ───────────────────────────────────────────────────────────

step "Verifying upload"

BLOB_COUNT=$(az storage blob list \
    --account-name "$ACCOUNT" \
    --container-name "$CONTAINER" \
    --auth-mode login \
    --query "length(@)" -o tsv 2>/dev/null)

detail "blobs in container: $BLOB_COUNT"
done_ "upload complete"

drummer_done "$ENDPOINT"
