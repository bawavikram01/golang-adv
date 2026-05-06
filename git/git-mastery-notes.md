# Git Mastery Notes — Zero to God Level

---

## 1. Foundations

### What is Git?
- **Distributed version control system** — every clone has full history
- No single point of failure; works offline

### The Three Areas
```
Working Directory  →  Staging Area (Index)  →  Repository (.git)
     (your files)        (git add)              (git commit)
```

### Essential Commands
| Command | Purpose |
|---------|---------|
| `git init` | Create new repo |
| `git clone <url>` | Clone existing repo |
| `git status` | See what's changed |
| `git add <file>` | Stage changes |
| `git add -p` | Stage interactively (hunk by hunk) |
| `git commit -m "msg"` | Commit staged changes |
| `git log --oneline --graph` | View history |
| `git diff` | Unstaged changes |
| `git diff --staged` | Staged changes |

---

## 2. Branching & Merging

### Branch Operations
| Command | Purpose |
|---------|---------|
| `git branch <name>` | Create branch |
| `git switch <name>` | Switch branch |
| `git switch -c <name>` | Create + switch |
| `git merge <branch>` | Merge into current |
| `git branch -d <branch>` | Delete merged branch |
| `git branch -D <branch>` | Force delete branch |

### Merge vs Rebase
- **Merge** — preserves history, creates merge commit
  ```bash
  git merge feature
  ```
- **Rebase** — replays commits on top of target, linear history
  ```bash
  git rebase main
  ```
- **Rule:** Never rebase commits already pushed to a shared branch

---

## 3. Working with Remotes

| Command | Purpose |
|---------|---------|
| `git remote -v` | List remotes |
| `git remote add origin <url>` | Add remote |
| `git fetch` | Download without merging |
| `git pull` | Fetch + merge |
| `git pull --rebase` | Fetch + rebase (cleaner) |
| `git push origin <branch>` | Push branch |
| `git push -u origin <branch>` | Push + set upstream |
| `git push origin --delete <branch>` | Delete remote branch |
| `git branch -vv` | See tracking relationships |

---

## 4. Undoing Things

### Quick Reference
| Command | Effect |
|---------|--------|
| `git restore --staged <file>` | Unstage a file |
| `git restore <file>` | Discard working dir changes |
| `git commit --amend` | Amend last commit |
| `git reset --soft HEAD~1` | Undo commit, keep staged |
| `git reset --mixed HEAD~1` | Undo commit, keep unstaged |
| `git reset --hard HEAD~1` | Undo commit, discard everything |
| `git revert <commit>` | New commit that undoes a past one (safe) |

### The Reset Spectrum
```
--soft   → moves HEAD only         → changes stay staged
--mixed  → moves HEAD + index      → changes in working dir (default)
--hard   → moves HEAD + index + WD → changes GONE
```

---

## 5. Stashing

| Command | Purpose |
|---------|---------|
| `git stash` | Stash tracked changes |
| `git stash -u` | Include untracked files |
| `git stash list` | List stashes |
| `git stash pop` | Apply + remove latest |
| `git stash apply stash@{2}` | Apply specific stash, keep it |
| `git stash drop stash@{0}` | Delete a stash |
| `git stash branch <name>` | Create branch from stash |

---

## 6. Interactive Rebase

```bash
git rebase -i HEAD~5    # Rewrite last 5 commits
```

### Editor Options
| Keyword | Action |
|---------|--------|
| `pick` | Keep commit as-is |
| `reword` | Change commit message |
| `edit` | Pause to amend the commit |
| `squash` | Meld into previous (keep message) |
| `fixup` | Meld into previous (discard message) |
| `drop` | Delete commit |

### Use Cases
- Clean messy commits before merging a PR
- Squash WIP commits into logical units
- Reorder or split commits
- Split a commit: use `edit` → `git reset HEAD~1` → re-commit in parts

---

## 7. Cherry-Pick, Bisect & Blame

### Cherry-Pick
```bash
git cherry-pick <commit-hash>    # Apply specific commit to current branch
```

### Bisect — Binary Search for Bugs
```bash
git bisect start
git bisect bad                   # Current is broken
git bisect good <known-good>     # This was fine
# Test middle commit, mark bad/good, repeat
git bisect reset                 # Done

# Automate with a script
git bisect run ./test.sh
```

### Blame
```bash
git blame <file>                 # Who wrote each line
git blame -L 10,20 <file>       # Lines 10–20 only
git log -p -S "search_term"     # Commits that added/removed a string (pickaxe)
```

---

## 8. Reflog — Your Safety Net

- Records **every HEAD movement** locally
- Even after `reset --hard`, you can recover
- Entries expire: 90 days (reachable), 30 days (unreachable)
- **Local only** — not shared with remotes

```bash
git reflog                       # Show reflog
git reset --hard HEAD@{3}        # Restore a previous state
git checkout HEAD@{5} -- file    # Recover a specific file
```

---

## 9. Configuration & Aliases

### Recommended Config
```bash
git config --global rerere.enabled true       # Remember conflict resolutions
git config --global pull.rebase true          # Always rebase on pull
git config --global fetch.prune true          # Auto-prune stale remote branches
git config --global diff.algorithm histogram  # Better diffs
```

### Power Aliases
```bash
git config --global alias.lg "log --oneline --graph --all --decorate"
git config --global alias.st "status -sb"
git config --global alias.unstage "restore --staged"
git config --global alias.last "log -1 HEAD --stat"
git config --global alias.amend "commit --amend --no-edit"
```

---

## 10. Worktrees, Submodules & Subtrees

### Worktrees — Multiple Working Dirs, One Repo
```bash
git worktree add ../hotfix hotfix-branch
git worktree list
git worktree remove ../hotfix
```

### Submodules — Repo Inside a Repo
```bash
git submodule add <url> path/
git submodule update --init --recursive
git submodule foreach git pull origin main
```

### Subtrees — Simpler Alternative to Submodules
```bash
git subtree add --prefix=lib/foo <url> main --squash
git subtree pull --prefix=lib/foo <url> main --squash
```

---

## 11. Git Internals — Object Model

### Object Types
| Object | Purpose |
|--------|---------|
| **blob** | File content (no filename) |
| **tree** | Directory listing (names → blobs/trees) |
| **commit** | Snapshot (tree + parent + metadata) |
| **tag** | Named pointer to a commit |

### Inspecting Objects
```bash
git cat-file -t <hash>       # Object type
git cat-file -p <hash>       # Pretty-print object
git ls-tree HEAD              # List tree at HEAD
git rev-parse HEAD            # Resolve ref to hash
git count-objects -vH         # Repo size stats
```

### Plumbing Commands (Manual Object Creation)
```bash
echo "hello" | git hash-object -w --stdin       # Write blob
git update-index --add --cacheinfo 100644 <hash> file.txt
git write-tree                                   # Write tree from index
git commit-tree <tree> -m "msg"                  # Write commit
```

### How Refs Work
```
.git/HEAD           → ref: refs/heads/main  (current branch)
.git/refs/heads/    → branch tips
.git/refs/tags/     → tags
.git/refs/remotes/  → remote tracking branches
.git/packed-refs    → optimized packed refs
```

---

## 12. Workflows & Strategies

### Commit Message Convention
```
<type>(<scope>): <subject>

<body>

<footer>
```
**Types:** `feat` `fix` `docs` `style` `refactor` `test` `chore`

### Branching Strategies
| Strategy | Description |
|----------|-------------|
| **GitHub Flow** | `main` + feature branches, deploy from main |
| **Git Flow** | `main` + `develop` + feature/release/hotfix branches |
| **Trunk-Based** | Short-lived branches, merge to main multiple times/day |

### Clean History Before Pushing
```bash
git fetch origin
git rebase origin/main              # Replay on latest main
git push --force-with-lease         # Safe force push
```
- `--force-with-lease` > `--force` — won't overwrite someone else's push

---

## 13. Performance & Maintenance

### Housekeeping
```bash
git gc                    # Garbage collect + pack objects
git prune                 # Remove unreachable objects
git fsck                  # Check repo integrity
git maintenance start     # Auto background maintenance (Git 2.29+)
```

### Large Repos
```bash
git clone --depth 1 <url>          # Shallow clone
git clone --filter=blob:none <url> # Blobless clone (on-demand blobs)
git sparse-checkout set src/       # Checkout only specific dirs
```

---

## 14. Git Hooks

Hooks live in `.git/hooks/` (or set `core.hooksPath`).

| Hook | Trigger |
|------|---------|
| `pre-commit` | Before commit (lint, format) |
| `commit-msg` | Validate commit message |
| `pre-push` | Before push (run tests) |
| `post-merge` | After merge (install deps) |
| `pre-rebase` | Before rebase |

**Tools for team hook management:** Husky (JS), pre-commit (Python)

---

## 15. Practice Exercises

1. Create a repo, make 5 commits, `rebase -i` to squash last 3
2. Create a merge conflict, resolve it, enable `rerere`, trigger it again
3. Use `git bisect` to find a bug-introducing commit
4. Do `reset --hard`, then recover with `reflog`
5. Explore `.git/` — read HEAD, inspect objects with `cat-file`
6. Write a `pre-commit` hook rejecting commits without a ticket number
7. `cherry-pick` across branches, handle conflicts
8. Use `worktrees` to work on two branches simultaneously

---

## Quick Cheat Sheet

```bash
# Daily workflow
git switch -c feature/xyz        # New branch
git add -p                       # Stage carefully
git commit                       # Commit
git fetch origin && git rebase origin/main  # Stay up to date
git push -u origin feature/xyz   # Push

# Oops recovery
git reflog                       # Find where you were
git reset --hard HEAD@{n}        # Go back

# Clean up before PR
git rebase -i origin/main        # Squash/reorder
git push --force-with-lease      # Safe force push

# Debugging
git bisect start                 # Find bad commit
git blame -L 10,20 file.py      # Who changed what
git log -p -S "bug_string"      # Search history
```
