package git

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
)

// Manager handles Git operations using the system's git CLI.
type Manager struct{}

// NewManager creates a new Git Manager.
func NewManager() *Manager {
	return &Manager{}
}

// execute runs a git command in the specified directory.
func (m *Manager) execute(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	// Disable interactive prompts to avoid hanging
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		msg := stderr.String()
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New(strings.TrimSpace(msg))
	}
	return strings.TrimSpace(out.String()), nil
}

// IsGitRepo checks if the given directory is within a Git repository.
func (m *Manager) IsGitRepo(dir string) bool {
	_, err := m.execute(dir, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

// Fetch updates the remote tracking branches.
func (m *Manager) Fetch(dir string) error {
	_, err := m.execute(dir, "fetch", "--all")
	return err
}

// GetCurrentBranch returns the name of the current branch.
func (m *Manager) GetCurrentBranch(dir string) (string, error) {
	return m.execute(dir, "rev-parse", "--abbrev-ref", "HEAD")
}

// GetMainBranch returns 'main' or 'master' depending on what exists locally.
func (m *Manager) GetMainBranch(dir string) (string, error) {
	branches, err := m.execute(dir, "branch", "--format=%(refname:short)")
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

// IsBehind checks if the local main branch is behind the remote tracking branch.
func (m *Manager) IsBehind(dir string) (bool, error) {
	if err := m.Fetch(dir); err != nil {
		return false, err
	}
	mainBranch, err := m.GetMainBranch(dir)
	if err != nil {
		return false, err
	}
	
	// Example: compare main with origin/main
	remoteTracking, err := m.execute(dir, "rev-parse", "--abbrev-ref", mainBranch+"@{upstream}")
	if err != nil {
		// No upstream configured, assume not behind
		return false, nil
	}
	
	out, err := m.execute(dir, "rev-list", "--left-right", "--count", mainBranch+"..."+remoteTracking)
	if err != nil {
		return false, err
	}
	
	parts := strings.Fields(out)
	if len(parts) == 2 {
		behindCount := parts[1]
		if behindCount != "0" {
			return true, nil
		}
	}
	return false, nil
}

// Pull updates the local repository.
func (m *Manager) Pull(dir string) error {
	_, err := m.execute(dir, "pull")
	return err
}

// ListBranches returns a list of local branches.
func (m *Manager) ListBranches(dir string) ([]string, error) {
	out, err := m.execute(dir, "branch", "--format=%(refname:short)")
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

// CreateBranch creates a new branch from a starting point (usually main).
func (m *Manager) CreateBranch(dir, name, from string) error {
	_, err := m.execute(dir, "checkout", "-b", name, from)
	return err
}

// PrepareBranch ensures the repository is on the target branch, creating it from main if necessary.
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

	// Check if branch exists
	branches, _ := m.ListBranches(dir)
	exists := false
	for _, b := range branches {
		if b == branchName {
			exists = true
			break
		}
	}

	if !exists {
		// Create from main
		_, err = m.execute(dir, "checkout", "-b", branchName, mainBranch)
	} else {
		// Switch to existing. Use --force if necessary to overcome local changes 
		// since Yapla is about to overwrite the file anyway.
		_, err = m.execute(dir, "checkout", branchName)
		if err != nil {
			// If normal checkout fails, try to stash or force? 
			// Let's try force checkout if it's just about overwritten files.
			_, err = m.execute(dir, "checkout", "-f", branchName)
		}
	}
	return err
}

// CommitAndPushChanges commits and pushes to the CURRENT branch.
func (m *Manager) CommitAndPushChanges(dir, branchName, message string) error {
	// Add changes
	_, err := m.execute(dir, "add", ".")
	if err != nil {
		return err
	}

	// Commit
	_, err = m.execute(dir, "commit", "-m", message)
	if err != nil {
		if strings.Contains(err.Error(), "nothing to commit") {
			return nil
		}
		return err
	}

	// Push
	_, err = m.execute(dir, "push", "-u", "origin", branchName)
	return err
}

// HasUncommittedChanges checks if the repository has uncommitted changes (dirty state).
func (m *Manager) HasUncommittedChanges(dir string) (bool, error) {
	out, err := m.execute(dir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(out)) > 0, nil
}

// DiscardChanges discards uncommitted changes.
func (m *Manager) DiscardChanges(dir string) error {
	_, err := m.execute(dir, "checkout", "--", ".")
	return err
}

// Checkout switches to a specific branch.
func (m *Manager) Checkout(dir, branch string) error {
	_, err := m.execute(dir, "checkout", branch)
	return err
}

// Clone clones a repository to the target directory.
func (m *Manager) Clone(url, targetDir string) error {
	_, err := m.execute("", "clone", url, targetDir)
	return err
}
