// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"solo/internal/tools"
	"strings"
	"time"
)

// GitResource represents a file or directory in a Git repository.
type GitResource struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	IsDirectory bool   `json:"isDirectory"`
	Format      string `json:"format"` // solo, postman, bruno, unknown
}

// GitLogEntry represents a single commit in the recent log.
type GitLogEntry struct {
	Hash    string `json:"hash"`
	Author  string `json:"author"`
	Message string `json:"message"`
	Date    string `json:"date"`
}

// CollectionStatus holds the full git status for a Git-backed collection directory.
type CollectionStatus struct {
	Branch             string        `json:"branch"`
	IsRebaseInProgress bool          `json:"isRebaseInProgress"`
	HasConflicts       bool          `json:"hasConflicts"`
	ConflictFiles      []string      `json:"conflictFiles"`
	StatusLines        []string      `json:"statusLines"`
	IsDirty            bool          `json:"isDirty"`
	Ahead              int           `json:"ahead"`
	Behind             int           `json:"behind"`
	RecentLog          []GitLogEntry `json:"recentLog"`
}

// Manager handles Git operations using the system's git CLI.
type Manager struct{}

// NewManager creates a new Git Manager.
func NewManager() *Manager {
	return &Manager{}
}

// IdentifyProvider attempts to identify the Git provider from the URL.
func (m *Manager) IdentifyProvider(url string) string {
	// Strip branch suffix if present for identification
	cleanUrl := strings.Split(url, "#")[0]
	cleanUrl = strings.ToLower(cleanUrl)

	if strings.Contains(cleanUrl, "github.com") {
		return "github"
	}
	if strings.Contains(cleanUrl, "gitlab.com") {
		return "gitlab"
	}
	if strings.Contains(cleanUrl, "bitbucket.org") {
		return "bitbucket"
	}
	if strings.Contains(cleanUrl, "dev.azure.com") || strings.Contains(cleanUrl, "visualstudio.com") {
		return "azure"
	}
	return "git" // Default generic Git provider
}

// GetRemoteBranches returns a list of branches in a remote Git repository.
func (m *Manager) GetRemoteBranches(url string) ([]string, error) {
	out, err := m.executeWithTimeout(20*time.Second, "", "ls-remote", "--heads", url)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(out, "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			ref := parts[1]
			branch := strings.TrimPrefix(ref, "refs/heads/")
			branches = append(branches, branch)
		}
	}
	return branches, nil
}

// BrowseRemote is kept for internal/future use but no longer used by the direct import flow.
func (m *Manager) BrowseRemote(url string) ([]GitResource, error) {
	return []GitResource{}, nil
}

// SetupGitCollection configures a local directory with a Git repository using sparse-checkout.
func (m *Manager) SetupGitCollection(url, remotePath, targetDir string) error {
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}

	// Parse URL and Branch (url#branch)
	gitUrl := url
	branch := ""
	if strings.Contains(url, "#") {
		parts := strings.Split(url, "#")
		gitUrl = parts[0]
		branch = parts[1]
	}

	slog.Info("Setting up Git collection", "url", gitUrl, "branch", branch, "remotePath", remotePath, "targetDir", absTargetDir)

	if err := os.MkdirAll(absTargetDir, 0755); err != nil {
		return err
	}

	// 1. Init
	if _, err := m.executeWithTimeout(10*time.Second, absTargetDir, "init"); err != nil {
		return err
	}

	// 2. Ensure remote is set correctly
	_, _ = m.executeWithTimeout(5*time.Second, absTargetDir, "remote", "remove", "origin")
	if _, err := m.executeWithTimeout(10*time.Second, absTargetDir, "remote", "add", "origin", gitUrl); err != nil {
		return err
	}

	// 3. Detect branch if not specified
	if branch == "" {
		out, err := m.executeWithTimeout(20*time.Second, absTargetDir, "ls-remote", "--symref", "origin", "HEAD")
		if err == nil {
			if strings.Contains(out, "refs/heads/") {
				parts := strings.Fields(out)
				for _, p := range parts {
					if strings.Contains(p, "refs/heads/") {
						branch = strings.TrimPrefix(p, "refs/heads/")
						break
					}
				}
			}
		}
		if branch == "" {
			branch = "main" // Fallback
		}
	}

	// 4. Configure sparse-checkout
	if _, err := m.executeWithTimeout(10*time.Second, absTargetDir, "sparse-checkout", "init"); err != nil {
		return err
	}
	if _, err := m.executeWithTimeout(10*time.Second, absTargetDir, "sparse-checkout", "set", "--skip-checks", remotePath); err != nil {
		return err
	}

	// 5. Fetch specifically the branch
	slog.Info("Fetching from Git", "branch", branch, "url", gitUrl)
	if _, err := m.executeWithTimeout(60*time.Second, absTargetDir, "fetch", "--depth", "1", "origin", branch); err != nil {
		// Fallback for master if main failed and was detected automatically
		if branch == "main" && !strings.Contains(url, "#") {
			if _, err2 := m.executeWithTimeout(60*time.Second, absTargetDir, "fetch", "--depth", "1", "origin", "master"); err2 == nil {
				branch = "master"
				err = nil
			}
		}
		if err != nil {
			return fmt.Errorf("failed to fetch from branch %s: %w", branch, err)
		}
	}

	// 6. Ensure local branch name matches remote branch name
	// Check if local branch exists, if not create it from origin/branch
	_, err = m.executeWithTimeout(5*time.Second, absTargetDir, "checkout", "-B", branch, "origin/"+branch)
	if err != nil {
		// Fallback: if origin/branch reference is not exactly there, try reset
		_, _ = m.executeWithTimeout(5*time.Second, absTargetDir, "checkout", "-b", branch)
		if _, err := m.executeWithTimeout(20*time.Second, absTargetDir, "reset", "--hard", "origin/"+branch); err != nil {
			return fmt.Errorf("failed to reset to branch %s: %w", branch, err)
		}
	}

	// 7. Set upstream tracking
	_, _ = m.executeWithTimeout(10*time.Second, absTargetDir, "branch", "--set-upstream-to=origin/"+branch, branch)

	return nil
}

// SyncGitCollection performs a pull --rebase, adds all changes, commits, and pushes.
func (m *Manager) SyncGitCollection(dir, filename, commitMessage string) error {
	branch, err := m.GetCurrentBranch(dir)
	if err != nil {
		branch = "main"
	}

	slog.Info("Syncing Git collection", "dir", dir, "branch", branch)

	// 1. Stage and commit local changes FIRST so the working tree is clean
	//    before rebase. Skip the commit entirely if there is nothing staged.
	if _, err := m.executeWithTimeout(15*time.Second, dir, "add", "."); err != nil {
		return fmt.Errorf("sync add failed: %w", err)
	}

	hasPending, _ := m.HasUncommittedChanges(dir)
	if hasPending {
		if commitMessage == "" {
			commitMessage = fmt.Sprintf("[%s] - Sync: %s - %s", tools.APP_NAME, filename, time.Now().Format(time.RFC3339))
		}
		if _, err := m.executeWithTimeout(15*time.Second, dir, "commit", "-m", commitMessage); err != nil {
			return fmt.Errorf("sync commit failed: %w", err)
		}
	}

	// 2. Fetch latest remote state
	if _, err := m.executeWithTimeout(60*time.Second, dir, "fetch", "origin", branch); err != nil {
		return fmt.Errorf("sync fetch failed: %w", err)
	}

	// 3. Rebase local commits on top of remote (working tree is clean now)
	if _, err := m.executeWithTimeout(60*time.Second, dir, "rebase", "origin/"+branch); err != nil {
		return fmt.Errorf("sync rebase failed (conflict?): %w", err)
	}

	// 4. Push — tolerate "Everything up-to-date" (exit 0) and
	//    "nothing to push" messages which are not real errors.
	_, pushErr := m.executeWithTimeout(60*time.Second, dir, "push", "origin", branch)
	if pushErr != nil {
		msg := pushErr.Error()
		if strings.Contains(msg, "Everything up-to-date") ||
			strings.Contains(msg, "up-to-date") {
			return nil
		}
		return fmt.Errorf("sync push failed: %w", pushErr)
	}
	return nil
}

func (m *Manager) executeWithTimeout(timeout time.Duration, dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o BatchMode=yes -o ConnectTimeout=15",
		"GIT_ASKPASS=echo",
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("git command timed out after %v", timeout)
		}
		// git writes some messages (e.g. "nothing to commit") to stdout, not stderr.
		// Prefer stderr, then fall back to stdout, then to the raw error.
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(out.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New(msg)
	}
	return strings.TrimSpace(out.String()), nil
}

func (m *Manager) execute(dir string, args ...string) (string, error) {
	return m.executeWithTimeout(60*time.Second, dir, args...)
}

func (m *Manager) GetCurrentBranch(dir string) (string, error) {
	return m.executeWithTimeout(10*time.Second, dir, "rev-parse", "--abbrev-ref", "HEAD")
}

func (m *Manager) IsGitRepo(dir string) bool {
	_, err := m.executeWithTimeout(10*time.Second, dir, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

func (m *Manager) Fetch(dir string) error {
	_, err := m.executeWithTimeout(60*time.Second, dir, "fetch", "--all")
	return err
}

func (m *Manager) GetMainBranch(dir string) (string, error) {
	branches, err := m.executeWithTimeout(10*time.Second, dir, "branch", "--format=%(refname:short)")
	if err != nil {
		return "", err
	}
	lines := strings.Split(branches, "\n")
	for _, b := range lines {
		if b == "main" || b == "master" {
			return b, nil
		}
	}
	return "", errors.New("no main or master branch found")
}

func (m *Manager) IsBehind(dir string) (bool, error) {
	if err := m.Fetch(dir); err != nil {
		return false, err
	}
	mainBranch, err := m.GetMainBranch(dir)
	if err != nil {
		return false, err
	}

	remoteTracking, err := m.executeWithTimeout(10*time.Second, dir, "rev-parse", "--abbrev-ref", mainBranch+"@{upstream}")
	if err != nil {
		return false, nil
	}

	out, err := m.executeWithTimeout(10*time.Second, dir, "rev-list", "--left-right", "--count", mainBranch+"..."+remoteTracking)
	if err != nil {
		return false, err
	}

	parts := strings.Fields(out)
	if len(parts) == 2 {
		if parts[1] != "0" {
			return true, nil
		}
	}
	return false, nil
}

func (m *Manager) Pull(dir string) error {
	branch, _ := m.GetCurrentBranch(dir)
	if branch == "" {
		branch = "main"
	}
	_, err := m.executeWithTimeout(60*time.Second, dir, "pull", "origin", branch)
	return err
}

func (m *Manager) ListBranches(dir string) ([]string, error) {
	out, err := m.executeWithTimeout(10*time.Second, dir, "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(out, "\n")
	var branches []string
	for _, l := range lines {
		if t := strings.TrimSpace(l); t != "" {
			branches = append(branches, t)
		}
	}
	return branches, nil
}

func (m *Manager) CreateBranch(dir, name, from string) error {
	_, err := m.executeWithTimeout(15*time.Second, dir, "checkout", "-b", name, from)
	return err
}

func (m *Manager) PrepareBranch(dir, branchName string) error {
	mainBranch, err := m.GetMainBranch(dir)
	if err != nil {
		return err
	}

	currentBranch, err := m.GetCurrentBranch(dir)
	if err != nil {
		return err
	}

	if currentBranch == branchName {
		return nil
	}

	branches, _ := m.ListBranches(dir)
	exists := false
	for _, b := range branches {
		if b == branchName {
			exists = true
			break
		}
	}

	if !exists {
		_, err = m.executeWithTimeout(15*time.Second, dir, "checkout", "-b", branchName, mainBranch)
	} else {
		_, err = m.executeWithTimeout(15*time.Second, dir, "checkout", branchName)
		if err != nil {
			_, err = m.executeWithTimeout(15*time.Second, dir, "checkout", "-f", branchName)
		}
	}
	return err
}

func (m *Manager) CommitAndPushChanges(dir, branchName, message string) error {
	_, err := m.executeWithTimeout(15*time.Second, dir, "add", ".")
	if err != nil {
		return err
	}

	_, err = m.executeWithTimeout(15*time.Second, dir, "commit", "-m", message)
	if err != nil && !strings.Contains(err.Error(), "nothing to commit") {
		return err
	}

	_, err = m.executeWithTimeout(60*time.Second, dir, "push", "-u", "origin", branchName)
	return err
}

func (m *Manager) HasUncommittedChanges(dir string) (bool, error) {
	out, err := m.executeWithTimeout(10*time.Second, dir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(out)) > 0, nil
}

func (m *Manager) DiscardChanges(dir string) error {
	_, err := m.executeWithTimeout(15*time.Second, dir, "checkout", "--", ".")
	return err
}

func (m *Manager) Checkout(dir, branch string) error {
	_, err := m.executeWithTimeout(15*time.Second, dir, "checkout", branch)
	return err
}

func (m *Manager) Clone(url, targetDir string) error {
	_, err := m.executeWithTimeout(60*time.Second, "", "clone", url, targetDir)
	return err
}

// ── Git Status ───────────────────────────────────────────────────────────────

// GetCollectionStatus returns a rich status snapshot of a local git repo directory.
func (m *Manager) GetCollectionStatus(dir string) (CollectionStatus, error) {
	status := CollectionStatus{}

	// Branch
	branch, err := m.GetCurrentBranch(dir)
	if err != nil {
		branch = "unknown"
	}
	status.Branch = branch

	// Rebase in progress: check for REBASE_HEAD or rebase-merge / rebase-apply dirs
	rebaseHead := filepath.Join(dir, ".git", "REBASE_HEAD")
	rebaseMerge := filepath.Join(dir, ".git", "rebase-merge")
	rebaseApply := filepath.Join(dir, ".git", "rebase-apply")
	if _, err := os.Stat(rebaseHead); err == nil {
		status.IsRebaseInProgress = true
	} else if _, err := os.Stat(rebaseMerge); err == nil {
		status.IsRebaseInProgress = true
	} else if _, err := os.Stat(rebaseApply); err == nil {
		status.IsRebaseInProgress = true
	}

	// Status lines + conflict detection
	out, err := m.executeWithTimeout(10*time.Second, dir, "status", "--short")
	if err == nil && out != "" {
		lines := strings.Split(out, "\n")
		for _, l := range lines {
			l = strings.TrimRight(l, "\r")
			if l == "" {
				continue
			}
			status.StatusLines = append(status.StatusLines, l)
			status.IsDirty = true
			xy := ""
			if len(l) >= 2 {
				xy = l[:2]
			}
			// Conflict codes: UU, AA, DD, AU, UA, DU, UD
			if xy == "UU" || xy == "AA" || xy == "DD" ||
				xy == "AU" || xy == "UA" || xy == "DU" || xy == "UD" {
				status.HasConflicts = true
				if len(l) > 3 {
					status.ConflictFiles = append(status.ConflictFiles, strings.TrimSpace(l[3:]))
				}
			}
		}
	}

	// Ahead / Behind (best-effort — may fail if no upstream)
	aheadBehind, err := m.executeWithTimeout(10*time.Second, dir,
		"rev-list", "--left-right", "--count", "@{upstream}...HEAD")
	if err == nil {
		parts := strings.Fields(aheadBehind)
		if len(parts) == 2 {
			fmt.Sscanf(parts[0], "%d", &status.Behind)
			fmt.Sscanf(parts[1], "%d", &status.Ahead)
		}
	}

	// Recent log (last 5 commits)
	logOut, err := m.executeWithTimeout(10*time.Second, dir,
		"log", "--oneline", "--format=%H|%an|%s|%cr", "-5")
	if err == nil && logOut != "" {
		for _, line := range strings.Split(logOut, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, "|", 4)
			entry := GitLogEntry{}
			if len(parts) > 0 && len(parts[0]) >= 7 {
				entry.Hash = parts[0][:7]
			}
			if len(parts) > 1 {
				entry.Author = parts[1]
			}
			if len(parts) > 2 {
				entry.Message = parts[2]
			}
			if len(parts) > 3 {
				entry.Date = parts[3]
			}
			status.RecentLog = append(status.RecentLog, entry)
		}
	}

	return status, nil
}

// ── Conflict resolution ──────────────────────────────────────────────────────

// KeepOurs resolves all conflicts by checking out our version, staging, and continuing the rebase.
func (m *Manager) KeepOurs(dir string) error {
	conflicts, err := m.conflictFiles(dir)
	if err != nil {
		return err
	}
	for _, f := range conflicts {
		if _, err := m.executeWithTimeout(15*time.Second, dir, "checkout", "--ours", f); err != nil {
			return fmt.Errorf("checkout --ours %s: %w", f, err)
		}
		if _, err := m.executeWithTimeout(10*time.Second, dir, "add", f); err != nil {
			return fmt.Errorf("add %s: %w", f, err)
		}
	}
	return m.continueOrCommitRebase(dir)
}

// KeepTheirs resolves all conflicts by checking out the remote version, staging, and continuing.
func (m *Manager) KeepTheirs(dir string) error {
	conflicts, err := m.conflictFiles(dir)
	if err != nil {
		return err
	}
	for _, f := range conflicts {
		if _, err := m.executeWithTimeout(15*time.Second, dir, "checkout", "--theirs", f); err != nil {
			return fmt.Errorf("checkout --theirs %s: %w", f, err)
		}
		if _, err := m.executeWithTimeout(10*time.Second, dir, "add", f); err != nil {
			return fmt.Errorf("add %s: %w", f, err)
		}
	}
	return m.continueOrCommitRebase(dir)
}

// AbortRebase aborts an in-progress rebase.
func (m *Manager) AbortRebase(dir string) error {
	_, err := m.executeWithTimeout(15*time.Second, dir, "rebase", "--abort")
	return err
}

// DiscardAllChanges discards all uncommitted local changes (tracked and untracked).
func (m *Manager) DiscardAllChanges(dir string) error {
	if _, err := m.executeWithTimeout(15*time.Second, dir, "checkout", "--", "."); err != nil {
		return err
	}
	_, err := m.executeWithTimeout(15*time.Second, dir, "clean", "-fd")
	return err
}

// conflictFiles returns the list of files with merge conflicts.
func (m *Manager) conflictFiles(dir string) ([]string, error) {
	out, err := m.executeWithTimeout(10*time.Second, dir, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil, err
	}
	var files []string
	for _, l := range strings.Split(out, "\n") {
		if f := strings.TrimSpace(l); f != "" {
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no conflict files found")
	}
	return files, nil
}

// continueOrCommitRebase attempts git rebase --continue; if not in a rebase it commits instead.
func (m *Manager) continueOrCommitRebase(dir string) error {
	// Set a dummy editor so rebase --continue doesn't open an interactive prompt
	env := append(os.Environ(),
		"GIT_EDITOR=true",
		"GIT_TERMINAL_PROMPT=0",
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rebase", "--continue")
	cmd.Dir = dir
	cmd.Env = env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		// If not in rebase, do a plain commit instead
		if strings.Contains(msg, "no rebase in progress") {
			_, err2 := m.executeWithTimeout(15*time.Second, dir, "commit",
				"-m", fmt.Sprintf("[%s] conflict resolved - %s", tools.APP_NAME, time.Now().Format(time.RFC3339)))
			return err2
		}
		return errors.New(msg)
	}
	return nil
}

// ── Open in Terminal ─────────────────────────────────────────────────────────

// OpenInTerminal opens the system's default terminal emulator at the given directory.
func (m *Manager) OpenInTerminal(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	slog.Info("Opening terminal", "dir", absDir, "os", runtime.GOOS)

	switch runtime.GOOS {
	case "darwin":
		return m.openTerminalMacOS(absDir)
	case "windows":
		return m.openTerminalWindows(absDir)
	default: // linux and others
		return m.openTerminalLinux(absDir)
	}
}

func (m *Manager) openTerminalMacOS(dir string) error {
	// Escape any single quotes in the path so the shell command inside
	// AppleScript handles paths with spaces (e.g. "Application Support").
	// Shell single-quote escaping: replace ' with '\''
	escapedDir := strings.ReplaceAll(dir, "'", `'\''`)

	// Try iTerm2 first, fall back to Terminal.app
	script := fmt.Sprintf(
		`tell application "iTerm2" to activate`+"\n"+
			`tell application "iTerm2"`+"\n"+
			`  tell current window to create tab with default profile`+"\n"+
			`  tell current session of current window to write text "cd '%s'"`+"\n"+
			`end tell`, escapedDir)

	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Fallback: Terminal.app
	script = fmt.Sprintf(
		`tell application "Terminal"`+"\n"+
			`  activate`+"\n"+
			`  do script "cd '%s'"`+"\n"+
			`end tell`, escapedDir)
	return exec.Command("osascript", "-e", script).Run()
}

func (m *Manager) openTerminalWindows(dir string) error {
	// Try Windows Terminal (wt.exe) first, fallback to cmd
	cmd := exec.Command("wt.exe", "-d", dir)
	if err := cmd.Start(); err == nil {
		return nil
	}
	// cmd /K "cd /d <path>" — quote the whole argument to handle spaces
	return exec.Command("cmd", "/c", "start", "cmd", "/K",
		fmt.Sprintf(`cd /d "%s"`, dir)).Start()
}

func (m *Manager) openTerminalLinux(dir string) error {
	// Escape single quotes for xterm's shell command
	escapedDir := strings.ReplaceAll(dir, "'", `'\''`)
	candidates := [][]string{
		{"gnome-terminal", "--working-directory=" + dir},
		{"konsole", "--workdir", dir},
		{"xfce4-terminal", "--working-directory=" + dir},
		{"kitty", "--directory", dir},
		{"alacritty", "--working-directory", dir},
		{"xterm", "-e", fmt.Sprintf("cd '%s' && exec $SHELL", escapedDir)},
	}
	for _, args := range candidates {
		cmd := exec.Command(args[0], args[1:]...)
		if err := cmd.Start(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no supported terminal emulator found on this system")
}
