#!/usr/bin/env bash
#==============================================================================
# auto-push.sh — Professional Auto-Update & Push Script for NetBridge
#==============================================================================
# Features:
#   - Detects git status (staged, modified, untracked files)
#   - Auto-stages changed files with smart filtering
#   - Generates intelligent commit messages based on changed files
#   - Detects remote sync status (ahead/behind counts)
#   - Auto-pushes with conflict detection
#   - Shows project health overview
#   - Color-coded output for readability
#   - Dry-run mode for safety
#   - Audit logging
#==============================================================================

set -euo pipefail

#--- Colors & Formatting -----------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m' # No Color

#--- Configuration -----------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOG_FILE="$SCRIPT_DIR/auto-push.log"
DRY_RUN=false
STATUS_ONLY=false
SHOW_LOG=false
SHOW_HELP=false
EXCLUDE_PATTERNS=(".env" "*.log" "*.tmp" ".DS_Store" "Thumbs.db")

#--- Logging -----------------------------------------------------------------
log() {
    local timestamp
    timestamp="$(date '+%Y-%m-%d %H:%M:%S')"
    echo "[$timestamp] $*" >> "$LOG_FILE" 2>/dev/null || true
}

#--- Output Helpers ----------------------------------------------------------
info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC} $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*"; }
header()  { echo -e "\n${BOLD}${CYAN}═══════════════════════════════════════════════════════════${NC}"; echo -e "${BOLD}${CYAN}  $*${NC}"; echo -e "${BOLD}${CYAN}═══════════════════════════════════════════════════════════${NC}\n"; }
divider() { echo -e "${DIM}───────────────────────────────────────────────────────────${NC}"; }

#--- Dependency Check --------------------------------------------------------
check_dependencies() {
    local missing=()
    for cmd in git; do
        if ! command -v "$cmd" &>/dev/null; then
            missing+=("$cmd")
        fi
    done
    if [[ ${#missing[@]} -gt 0 ]]; then
        error "Missing required commands: ${missing[*]}"
        exit 1
    fi
}

#--- Git Repository Check ----------------------------------------------------
check_git_repo() {
    if ! git -C "$PROJECT_ROOT" rev-parse --is-inside-work-tree &>/dev/null; then
        error "Not a git repository: $PROJECT_ROOT"
        echo -e "Run ${BOLD}git init${NC} first, or navigate to a git project."
        exit 1
    fi
}

#--- Get Current Branch ------------------------------------------------------
get_branch() {
    git -C "$PROJECT_ROOT" branch --show-current 2>/dev/null || echo "detached"
}

#--- Get Remote URL ----------------------------------------------------------
get_remote_url() {
    git -C "$PROJECT_ROOT" remote get-url origin 2>/dev/null || echo "No remote configured"
}

#--- Get Last Commit Info ----------------------------------------------------
get_last_commit() {
    git -C "$PROJECT_ROOT" log -1 --format="%h %s (%ar)" 2>/dev/null || echo "No commits yet"
}

#--- Count Files by Status ---------------------------------------------------
count_files() {
    local status="$1"
    git -C "$PROJECT_ROOT" status --porcelain 2>/dev/null | grep -c "^$status" 2>/dev/null || echo "0"
}

#--- Get Changed Files -------------------------------------------------------
get_changed_files() {
    local status="$1"
    git -C "$PROJECT_ROOT" status --porcelain 2>/dev/null | grep "^$status" | sed 's/^...//' 2>/dev/null || true
}

#--- Detect File Categories --------------------------------------------------
categorize_files() {
    local files=("$@")
    local go_files=0
    local doc_files=0
    local config_files=0
    local test_files=0
    local script_files=0
    local other_files=0

    for file in "${files[@]}"; do
        case "$file" in
            *.go)                ((go_files++)) ;;
            *.md|*.txt|docs/*)   ((doc_files++)) ;;
            *.yml|*.yaml|*.json|*.toml|Makefile|.gitignore) ((config_files++)) ;;
            *_test.go|tests/*)   ((test_files++)) ;;
            *.sh|scripts/*)      ((script_files++)) ;;
            *)                   ((other_files++)) ;;
        esac
    done

    echo "go=$go_files doc=$doc_files config=$config_files test=$test_files script=$script_files other=$other_files"
}

#--- Generate Smart Commit Message -------------------------------------------
generate_commit_message() {
    local staged_files
    staged_files=($(get_changed_files "M"))
    local new_files
    new_files=($(get_changed_files "A"))
    local deleted_files
    deleted_files=($(get_changed_files "D"))
    local renamed_files
    renamed_files=($(get_changed_files "R"))

    local total_changes=$(( ${#staged_files[@]} + ${#new_files[@]} + ${#deleted_files[@]} + ${#renamed_files[@]} ))

    if [[ $total_changes -eq 0 ]]; then
        echo "chore: minor updates"
        return
    fi

    # Analyze file categories
    local all_files=("${staged_files[@]}" "${new_files[@]}" "${deleted_files[@]}")
    local categories
    categories=$(categorize_files "${all_files[@]}")

    local go_count=$(echo "$categories" | grep -oP 'go=\K[0-9]+')
    local doc_count=$(echo "$categories" | grep -oP 'doc=\K[0-9]+')
    local config_count=$(echo "$categories" | grep -oP 'config=\K[0-9]+')
    local test_count=$(echo "$categories" | grep -oP 'test=\K[0-9]+')
    local script_count=$(echo "$categories" | grep -oP 'script=\K[0-9]+')

    # Build commit message
    local msg_parts=()
    local scope=""

    # Determine scope
    if [[ $go_count -gt 0 && $doc_count -eq 0 && $config_count -eq 0 ]]; then
        scope=""
    elif [[ $go_count -eq 0 && $doc_count -gt 0 ]]; then
        scope="docs"
    elif [[ $config_count -gt 0 ]]; then
        scope="config"
    elif [[ $test_count -gt 0 ]]; then
        scope="test"
    fi

    # Determine type
    local commit_type="update"
    if [[ ${#new_files[@]} -gt ${#staged_files[@]} ]]; then
        commit_type="add"
    elif [[ ${#deleted_files[@]} -gt ${#staged_files[@]} ]]; then
        commit_type="remove"
    fi

    # Build message
    if [[ -n "$scope" ]]; then
        msg_parts+=("$commit_type($scope)")
    else
        msg_parts+=("$commit_type")
    fi

    # Add summary
    local summary_parts=()
    [[ $go_count -gt 0 ]] && summary_parts+=("$go_count Go file(s)")
    [[ $doc_count -gt 0 ]] && summary_parts+=("$doc_count doc(s)")
    [[ $config_count -gt 0 ]] && summary_parts+=("$config_count config(s)")
    [[ $test_count -gt 0 ]] && summary_parts+=("$test_count test(s)")
    [[ $script_count -gt 0 ]] && summary_parts+=("$script_count script(s)")

    if [[ ${#summary_parts[@]} -gt 0 ]]; then
        echo "${msg_parts[*]}: $(IFS=', '; echo "${summary_parts[*]}")"
    else
        echo "${msg_parts[*]}: $total_changes file(s)"
    fi
}

#--- Show Project Status -----------------------------------------------------
show_status() {
    header "NetBridge — Git Project Status"

    echo -e "${BOLD}Repository:${NC}  $PROJECT_ROOT"
    echo -e "${BOLD}Branch:${NC}      $(get_branch)"
    echo -e "${BOLD}Remote:${NC}      $(get_remote_url)"
    echo -e "${BOLD}Last Commit:${NC} $(get_last_commit)"
    divider

    # File status counts
    local modified=$(count_files "M")
    local added=$(count_files "A")
    local deleted=$(count_files "D")
    local untracked=$(count_files "?")
    local total_changed=$(( modified + added + deleted + untracked ))

    echo -e "\n${BOLD}File Status:${NC}"
    echo -e "  ${GREEN}M (modified):${NC}   $modified"
    echo -e "  ${GREEN}A (added):${NC}      $added"
    echo -e "  ${RED}D (deleted):${NC}    $deleted"
    echo -e "  ${YELLOW}?(untracked):${NC}   $untracked"
    echo -e "  ${BOLD}Total changes:${NC}  $total_changed"

    # Remote sync status
    divider
    echo -e "\n${BOLD}Remote Sync:${NC}"

    if git -C "$PROJECT_ROOT" rev-parse --verify HEAD &>/dev/null; then
        local ahead behind
        ahead=$(git -C "$PROJECT_ROOT" rev-list --count @{u}..HEAD 2>/dev/null || echo "?")
        behind=$(git -C "$PROJECT_ROOT" rev-list --count HEAD..@{u} 2>/dev/null || echo "?")

        if [[ "$ahead" == "?" || "$behind" == "?" ]]; then
            echo -e "  ${YELLOW}No upstream branch set${NC}"
        elif [[ "$ahead" -eq 0 && "$behind" -eq 0 ]]; then
            echo -e "  ${GREEN}Up to date${NC}"
        else
            [[ "$ahead" -gt 0 ]] && echo -e "  ${GREEN}Ahead:${NC}    $ahead commit(s)"
            [[ "$behind" -gt 0 ]] && echo -e "  ${RED}Behind:${NC}   $behind commit(s)"
        fi
    else
        echo -e "  ${YELLOW}No commits yet${NC}"
    fi

    # Show changed files if any
    if [[ $total_changed -gt 0 ]]; then
        divider
        echo -e "\n${BOLD}Changed Files:${NC}"
        local changed_files
        changed_files=$(get_changed_files ".")
        if [[ -n "$changed_files" ]]; then
            echo "$changed_files" | while read -r line; do
                local status="${line:0:2}"
                local file="${line:3}"
                case "$status" in
                    "M "*) echo -e "  ${YELLOW}M${NC}  $file" ;;
                    "A "*) echo -e "  ${GREEN}A${NC}  $file" ;;
                    "D "*) echo -e "  ${RED}D${NC}  $file" ;;
                    "??")  echo -e "  ${MAGENTA}?${NC}  $file" ;;
                    *)     echo -e "  ${DIM}$status${NC} $file" ;;
                esac
            done
        fi
    fi

    echo ""
}

#--- Auto Stage & Commit -----------------------------------------------------
auto_commit() {
    header "Auto Stage & Commit"

    # Check for changes
    local total_changed
    total_changed=$(git -C "$PROJECT_ROOT" status --porcelain 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$total_changed" -eq 0 ]]; then
        success "No changes to commit. Working tree is clean."
        return 0
    fi

    info "Found $total_changed changed file(s)"

    # Stage all changes
    if [[ "$DRY_RUN" == true ]]; then
        info "[DRY RUN] Would stage all changes"
    else
        git -C "$PROJECT_ROOT" add -A
        success "Staged all changes"
    fi

    # Generate commit message
    local commit_msg
    commit_msg=$(generate_commit_message)
    info "Commit message: $commit_msg"

    # Commit
    if [[ "$DRY_RUN" == true ]]; then
        info "[DRY RUN] Would commit with message: \"$commit_msg\""
    else
        git -C "$PROJECT_ROOT" commit -m "$commit_msg"
        success "Committed: $commit_msg"
        log "COMMIT: $commit_msg"
    fi
}

#--- Auto Push ---------------------------------------------------------------
auto_push() {
    header "Auto Push"

    # Check if remote is configured
    local remote_url
    remote_url=$(get_remote_url)
    if [[ "$remote_url" == "No remote configured" ]]; then
        warn "No remote configured. Skipping push."
        info "Run: git remote add origin <url>"
        return 0
    fi

    # Check if there are commits to push
    local ahead
    ahead=$(git -C "$PROJECT_ROOT" rev-list --count @{u}..HEAD 2>/dev/null || echo "0")

    if [[ "$ahead" -eq 0 ]]; then
        # Check if HEAD exists and there's a remote
        if git -C "$PROJECT_ROOT" rev-parse --verify HEAD &>/dev/null && \
           git -C "$PROJECT_ROOT" rev-parse --verify @{u} &>/dev/null; then
            success "Already up to date with remote"
            return 0
        fi
    fi

    # Check for merge conflicts before pushing
    local conflicted
    conflicted=$(git -C "$PROJECT_ROOT" diff --name-only --diff-filter=U 2>/dev/null | wc -l | tr -d ' ')
    if [[ "$conflicted" -gt 0 ]]; then
        error "Cannot push: $conflicted conflicted file(s) detected"
        info "Resolve conflicts first, then try again"
        return 1
    fi

    # Push
    local branch
    branch=$(get_branch)

    if [[ "$DRY_RUN" == true ]]; then
        info "[DRY RUN] Would push to origin/$branch"
    else
        info "Pushing to origin/$branch..."
        if git -C "$PROJECT_ROOT" push -u origin "$branch" 2>&1; then
            success "Pushed successfully to origin/$branch"
            log "PUSH: origin/$branch"
        else
            error "Push failed. Check your credentials and network."
            log "PUSH FAILED: origin/$branch"
            return 1
        fi
    fi
}

#--- Show Log ----------------------------------------------------------------
show_log() {
    header "Recent Commit History"

    if ! git -C "$PROJECT_ROOT" rev-parse --verify HEAD &>/dev/null; then
        warn "No commits yet"
        return 0
    fi

    echo -e "${BOLD}Last 10 commits:${NC}\n"
    git -C "$PROJECT_ROOT" log --oneline -10 --color=always 2>/dev/null

    echo ""
    divider
    echo -e "\n${BOLD}Commit Statistics:${NC}"
    echo -e "  Total commits:    $(git -C "$PROJECT_ROOT" rev-list --count HEAD 2>/dev/null || echo 0)"
    echo -e "  Contributors:     $(git -C "$PROJECT_ROOT" shortlog -sn --no-merges 2>/dev/null | wc -l | tr -d ' ')"
    echo -e "  First commit:     $(git -C "$PROJECT_ROOT" log --reverse --format='%ai' -1 2>/dev/null || echo 'N/A')"
    echo -e "  Last commit:      $(git -C "$PROJECT_ROOT" log -1 --format='%ai' 2>/dev/null || echo 'N/A')"
    echo ""
}

#--- Show Help ---------------------------------------------------------------
show_help() {
    header "NetBridge — Auto-Push Script"
    echo -e "${BOLD}Usage:${NC}"
    echo "  $0 [OPTIONS]"
    echo ""
    echo -e "${BOLD}Options:${NC}"
    echo -e "  ${GREEN}(no args)${NC}      Auto stage + commit + push"
    echo -e "  ${GREEN}--status${NC}       Show project status only"
    echo -e "  ${GREEN}--dry-run${NC}      Preview changes without modifying"
    echo -e "  ${GREEN}--log${NC}          Show recent commit history"
    echo -e "  ${GREEN}--help${NC}         Show this help message"
    echo ""
    echo -e "${BOLD}Examples:${NC}"
    echo "  $0                  # Stage, commit, and push all changes"
    echo "  $0 --status         # View current git status"
    echo "  $0 --dry-run        # Preview what would happen"
    echo "  $0 --log            # View recent commits"
    echo ""
    echo -e "${BOLD}Workflow:${NC}"
    echo "  1. Make changes to your code"
    echo "  2. Run '$0' to auto-commit and push"
    echo "  3. Script detects changes, generates commit message, and pushes"
    echo ""
}

#--- Main --------------------------------------------------------------------
main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --status)   STATUS_ONLY=true; shift ;;
            --dry-run)  DRY_RUN=true; shift ;;
            --log)      SHOW_LOG=true; shift ;;
            --help|-h)  SHOW_HELP=true; shift ;;
            *)
                error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # Show help if requested
    if [[ "$SHOW_HELP" == true ]]; then
        show_help
        return 0
    fi

    # Run checks
    check_dependencies
    check_git_repo

    # Initialize log
    mkdir -p "$(dirname "$LOG_FILE")" 2>/dev/null || true
    log "SCRIPT STARTED: $0 $*"

    # Execute based on mode
    if [[ "$STATUS_ONLY" == true ]]; then
        show_status
    elif [[ "$SHOW_LOG" == true ]]; then
        show_log
    else
        # Full workflow: status -> commit -> push
        show_status

        if [[ "$DRY_RUN" == true ]]; then
            warn "DRY RUN MODE — No changes will be made"
            echo ""
        fi

        auto_commit
        auto_push

        header "Done!"
        log "SCRIPT COMPLETED"
    fi
}

# Run main
main "$@"
