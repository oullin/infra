#!/bin/bash
set -euo pipefail

# Marks all unviewed files in a GitHub pull request as "Viewed" using the GraphQL API.
#
# Usage:
#   GITHUB_TOKEN="ghp_xxx" bash scripts/mark_pr_files_viewed.sh --owner=<name> --repo=<repo> --pr=<number>

BATCH_SIZE=20

# --- Colors & symbols ---
BOLD="\033[1m"
DIM="\033[2m"
GREEN="\033[32m"
YELLOW="\033[33m"
RED="\033[31m"
CYAN="\033[36m"
RESET="\033[0m"

info()    { echo -e "  ${CYAN}>${RESET} $1"; }
success() { echo -e "  ${GREEN}✔${RESET} $1"; }
skip()    { echo -e "  ${DIM}–${RESET} ${DIM}$1${RESET}"; }
warn()    { echo -e "  ${YELLOW}!${RESET} $1"; }
error()   { echo -e "  ${RED}✖${RESET} $1" >&2; }
header()  { echo -e "\n${BOLD}$1${RESET}"; }

# --- Help ---
show_help() {
  echo ""
  echo -e "${BOLD}Mark PR Files as Viewed${RESET}"
  echo -e "${DIM}Marks all unviewed files in a GitHub pull request as \"Viewed\"${RESET}"
  echo -e "${DIM}using the GraphQL API.${RESET}"
  echo ""
  echo -e "${BOLD}Usage:${RESET}"
  echo -e "  $0 ${CYAN}--owner${RESET}=<name> ${CYAN}--repo${RESET}=<repo> ${CYAN}--pr${RESET}=<number> [${CYAN}--token${RESET}=<ref>]"
  echo ""
  echo -e "${BOLD}Options:${RESET}"
  echo -e "  ${CYAN}--owner${RESET}=<name>     GitHub repository owner"
  echo -e "  ${CYAN}--repo${RESET}=<repo>      GitHub repository name"
  echo -e "  ${CYAN}--pr${RESET}=<number>      Pull request number"
  echo -e "  ${CYAN}--token${RESET}=<ref>      1Password secret reference (vault/item/field)"
  echo -e "  ${CYAN}--help${RESET}             Show this help message"
  echo ""
  echo -e "${BOLD}Authentication (one of):${RESET}"
  echo -e "  ${CYAN}--token${RESET}            Read token from 1Password via ${DIM}op read${RESET}"
  echo -e "  ${CYAN}GITHUB_TOKEN${RESET}       Environment variable with a fine-grained PAT"
  echo ""
  echo -e "${BOLD}Examples:${RESET}"
  echo -e "  ${DIM}# Using 1Password${RESET}"
  echo -e "  bash $0 --owner=oullin --repo=infra --pr=11 --token=Private/GitHub\\ PAT/token"
  echo ""
  echo -e "  ${DIM}# Using an environment variable${RESET}"
  echo -e "  GITHUB_TOKEN=\"ghp_xxx\" bash $0 --owner=oullin --repo=infra --pr=11"
  echo ""
}

# --- Parse named arguments ---
OWNER=""
REPO=""
PR_NUMBER=""
TOKEN_REF=""

for arg in "$@"; do
  case "$arg" in
    --help|-h)  show_help; exit 0 ;;
    --owner=*)  OWNER="${arg#*=}" ;;
    --repo=*)   REPO="${arg#*=}" ;;
    --pr=*)     PR_NUMBER="${arg#*=}" ;;
    --token=*)  TOKEN_REF="${arg#*=}" ;;
    *)
      error "Unknown argument: $arg"
      echo -e "  Run ${CYAN}$0 --help${RESET} for usage information." >&2
      exit 1
      ;;
  esac
done

# --- Resolve token ---
if [[ -n "$TOKEN_REF" ]]; then
  if ! command -v op &> /dev/null; then
    error "'op' CLI is not installed. Install it from https://1password.com/downloads/command-line"
    exit 1
  fi
  info "Reading token from 1Password (${DIM}${TOKEN_REF}${RESET})"
  GITHUB_TOKEN=$(op read "op://${TOKEN_REF}")
fi

if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  error "No token provided. Use ${CYAN}--token${RESET} or set ${CYAN}GITHUB_TOKEN${RESET}."
  echo -e "  Run ${CYAN}$0 --help${RESET} for usage information." >&2
  exit 1
fi

if [[ -z "$OWNER" || -z "$REPO" || -z "$PR_NUMBER" ]]; then
  error "Missing required arguments."
  echo -e "  Run ${CYAN}$0 --help${RESET} for usage information." >&2
  exit 1
fi

if ! command -v gh &> /dev/null; then
  error "'gh' CLI is not installed. Install it from https://cli.github.com"
  exit 1
fi

export GH_TOKEN="$GITHUB_TOKEN"

# --- Fetch PR files and their viewed state via GraphQL ---
header "Fetching PR #${PR_NUMBER} from ${OWNER}/${REPO}"

CURSOR=""
PR_NODE_ID=""
ALL_UNVIEWED=()
ALREADY_VIEWED=0

while true; do
  CURSOR_ARG="null"
  if [[ -n "$CURSOR" ]]; then
    CURSOR_ARG="\"$CURSOR\""
  fi

  RESPONSE=$(gh api graphql -f query="
    query {
      repository(owner: \"${OWNER}\", name: \"${REPO}\") {
        pullRequest(number: ${PR_NUMBER}) {
          id
          files(first: 100, after: ${CURSOR_ARG}) {
            pageInfo { hasNextPage endCursor }
            nodes {
              path
              viewerViewedState
            }
          }
        }
      }
    }
  ")

  if [[ -z "$PR_NODE_ID" ]]; then
    PR_NODE_ID=$(echo "$RESPONSE" | gh api graphql -f query='query { __typename }' > /dev/null 2>&1 || true)
    PR_NODE_ID=$(echo "$RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['repository']['pullRequest']['id'])")
  fi

  while IFS=$'\t' read -r path state; do
    if [[ "$state" == "VIEWED" ]]; then
      ALREADY_VIEWED=$((ALREADY_VIEWED + 1))
    else
      ALL_UNVIEWED+=("$path")
    fi
  done < <(echo "$RESPONSE" | python3 -c "
import sys, json
nodes = json.load(sys.stdin)['data']['repository']['pullRequest']['files']['nodes']
for n in nodes:
    print(n['path'] + '\t' + n['viewerViewedState'])
")

  HAS_NEXT=$(echo "$RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['repository']['pullRequest']['files']['pageInfo']['hasNextPage'])")
  if [[ "$HAS_NEXT" == "True" ]]; then
    CURSOR=$(echo "$RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['repository']['pullRequest']['files']['pageInfo']['endCursor'])")
  else
    break
  fi
done

TOTAL_FILES=$((${#ALL_UNVIEWED[@]} + ALREADY_VIEWED))

info "Found ${TOTAL_FILES} file(s) in PR"

if [[ $ALREADY_VIEWED -gt 0 ]]; then
  skip "${ALREADY_VIEWED} file(s) already viewed"
fi

if [[ ${#ALL_UNVIEWED[@]} -eq 0 ]]; then
  echo ""
  success "All files are already viewed. Nothing to do."
  exit 0
fi

info "${#ALL_UNVIEWED[@]} file(s) to mark as viewed"

# --- Mark files as viewed in batches ---
header "Marking files as viewed (batch size: ${BATCH_SIZE})"

MARKED=0
UNVIEWED_COUNT=${#ALL_UNVIEWED[@]}

for ((i = 0; i < UNVIEWED_COUNT; i += BATCH_SIZE)); do
  BATCH=("${ALL_UNVIEWED[@]:i:BATCH_SIZE}")
  BATCH_NUM=$(( (i / BATCH_SIZE) + 1 ))
  TOTAL_BATCHES=$(( (UNVIEWED_COUNT + BATCH_SIZE - 1) / BATCH_SIZE ))

  # Build a single GraphQL mutation with aliased fields for the batch.
  MUTATION="mutation {"
  for j in "${!BATCH[@]}"; do
    FILE="${BATCH[$j]}"
    ESCAPED_FILE=$(echo "$FILE" | sed 's/\\/\\\\/g; s/"/\\"/g')
    MUTATION+="
    f${j}: markFileAsViewed(input: { pullRequestId: \"${PR_NODE_ID}\", path: \"${ESCAPED_FILE}\" }) {
      pullRequest { id }
    }"
  done
  MUTATION+="
  }"

  gh api graphql -f query="$MUTATION" > /dev/null

  for FILE in "${BATCH[@]}"; do
    MARKED=$((MARKED + 1))
    success "${FILE}"
  done

  if [[ $TOTAL_BATCHES -gt 1 ]]; then
    info "Batch ${BATCH_NUM}/${TOTAL_BATCHES} complete"
  fi
done

# --- Summary ---
header "Summary"
echo -e "  ${GREEN}${MARKED}${RESET} file(s) marked as viewed"
if [[ $ALREADY_VIEWED -gt 0 ]]; then
  echo -e "  ${DIM}${ALREADY_VIEWED} file(s) were already viewed${RESET}"
fi
echo ""
