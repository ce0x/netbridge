# Plan: Push NetBridge to GitHub & Run on Linux

## Execution Mode
**I will do it for you.** Steps 2-4 (create .gitignore, init git, stage, commit, push) will be executed automatically. Step 1 (create repo on GitHub) and Step 5 (run on Linux) are manual.

## Overview
NetBridge is a Go project (module: `github.com/netbridge/netbridge`) that needs to be pushed to GitHub and run on a Linux server. The project already has a Makefile, `.goreleaser.yml`, systemd service file, and build scripts — it's well-structured for this.

---

## Step 1: Create GitHub Repository (Manual — user does this)

- GitHub: `https://github.com/ce0x/`
- Repo name: `netbridge`
- Full URL: `https://github.com/ce0x/netbridge`

1. Go to [github.com/new](https://github.com/new)
2. Repository name: `netbridge`
3. Set visibility to **Private**
4. **Do NOT** initialize with README, .gitignore, or license
5. Click "Create repository"

---

## Step 2: Initialize Git & Create .gitignore

In project root:

```bash
git init
```

Create `.gitignore` file:
```gitignore
# Binaries
build/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of go coverage
*.out
*.prof

# Dependency directories
vendor/

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.local
```

**Note:** `go.sum` is NOT excluded — it's needed for reproducible builds.

---

## Step 3: Stage & Commit

```bash
git add .
git commit -m "Initial commit: NetBridge network toolkit"
```

---

## Step 4: Push to GitHub

```bash
git remote add origin https://github.com/ce0x/netbridge.git
git branch -M main
git push -u origin main
```

---

## Step 5: Run on Linux Server

### 5a. Clone & Build
```bash
git clone https://github.com/ce0x/netbridge.git
cd netbridge

# Install Go 1.22+ if not present
sudo apt update && sudo apt install -y golang-go

# Build
make build
# Binary will be at: build/netbridge
```

### 5b. Install System-wide
```bash
sudo make install
# Installs to /usr/local/bin/netbridge
```

### 5c. Run as systemd Service (optional)
```bash
sudo cp systemd/netbridge.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable netbridge
sudo systemctl start netbridge
```

### 5d. Quick Test
```bash
netbridge --help
netbridge status
```

---

## Step 6: Share with Others

Others can install via:
```bash
git clone https://github.com/ce0x/netbridge.git
cd netbridge
make build
sudo make install
```

Or download pre-built binaries from GitHub Releases (if using goreleaser):
```bash
goreleaser release --clean
```

---

## Step 7: Auto-Push Script

Use `auto-push.sh` for automated git workflow:
```bash
./auto-push.sh              # Auto stage + commit + push
./auto-push.sh --status     # Show project status only
./auto-push.sh --dry-run    # Preview without changes
./auto-push.sh --log        # Show commit history
./auto-push.sh --help       # Show all options
```

---

## Files Created
| File | Purpose |
|------|---------|
| `.gitignore` | Exclude build artifacts, IDE files, OS files |
| `auto pushing git/plan.md` | This deployment plan |
| `auto pushing git/auto-push.sh` | Professional auto-update & push script |

## Files Already Present
- `go.mod`, `go.sum` — Go module files
- `Makefile` — Build automation
- `.goreleaser.yml` — Release automation
- `systemd/netbridge.service` — Service file
- `scripts/*.sh` — Build/install scripts

---

## Verification
1. After pushing: visit `https://github.com/ce0x/netbridge` — confirm files are visible
2. On Linux: clone, `make build`, run `./build/netbridge --help` — confirm binary works
3. Test `auto-push.sh --status` — confirm it detects project state correctly
4. Test `auto-push.sh --dry-run` — confirm it previews changes without modifying
5. Test `auto-push.sh` — confirm full auto-push workflow works
