package main

import (
	"sort"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// check if current dir is a git repo
func isGitRepo() bool {
	_, err := git.PlainOpen(".")
	return err == nil
}

// initialize a new git repository
func initGitRepo() error {
	_, err := git.PlainInit(".", false)
	return err
}

// initialize a new git repository with custom default branch
func initGitRepoWithBranch(defaultBranch string) error {
	r, err := git.PlainInit(".", false)
	if err != nil {
		return err
	}

	// create and checkout the default branch
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	// create an initial empty commit on the default branch
	_, err = w.Commit("initial commit", &git.CommitOptions{})
	if err != nil {
		return err
	}

	if defaultBranch != "master" {
		head, err := r.Head()
		if err != nil {
			return err
		}

		newBranchRef := plumbing.NewBranchReferenceName(defaultBranch)
		err = r.Storer.SetReference(plumbing.NewHashReference(newBranchRef, head.Hash()))
		if err != nil {
			return err
		}

		err = r.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, newBranchRef))
		if err != nil {
			return err
		}
	}

	return nil
}

// add remote origin to the repository
func addRemoteOrigin(url string) error {
	r, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	// convert https url to ssh url for github
	// ideally this would be user choice ig
	sshURL := convertToSSHURL(url)

	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{sshURL},
	})
	return err
}

// converts a github https to ssh url
func convertToSSHURL(httpsURL string) string {
	// https://github.com/<user/org>/<repo>.git
	// git@github.com:<user/org>/<repo>.git

	if strings.Contains(httpsURL, "github.com") && strings.HasPrefix(httpsURL, "https://") {
		parts := strings.Split(httpsURL, "github.com/")
		if len(parts) == 2 {
			repoPath := strings.TrimSuffix(parts[1], ".git")
			return "git@github.com:" + repoPath + ".git"
		}
	}

	return httpsURL
}

// retrieve git status
func getGitStatus() ([]FileStatus, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return nil, err
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	status, err := w.Status()
	if err != nil {
		return nil, err
	}

	// create a map to track all files and their statuses
	fileMap := make(map[string]FileStatus)

	for path, fileStatus := range status {
		var fileStatusStr string

		// determine the overall status for the file
		if fileStatus.Staging == git.Added || fileStatus.Staging == git.Modified || fileStatus.Staging == git.Deleted {
			fileStatusStr = "staged"
		} else if fileStatus.Worktree == git.Modified || fileStatus.Worktree == git.Deleted {
			fileStatusStr = "unstaged"
		} else if fileStatus.Worktree == git.Untracked {
			fileStatusStr = "untracked"
		} else {
			continue
		}

		fileMap[path] = FileStatus{
			Path:     path,
			Status:   fileStatusStr,
			Selected: false,
		}
	}

	var files []FileStatus
	for _, file := range fileMap {
		files = append(files, file)
	}

	sortFiles(files)

	return files, nil
}

func sortFiles(files []FileStatus) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}

func openRepo() (*git.Repository, *git.Worktree, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return nil, nil, err
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, nil, err
	}

	return r, w, nil
}

func stageFile(path string) error {
	_, w, err := openRepo()
	if err != nil {
		return err
	}

	_, err = w.Add(path)
	return err
}

func unstageFile(path string) error {
	r, w, err := openRepo()
	if err != nil {
		return err
	}

	// check if HEAD exists
	_, err = r.Head()
	if err != nil {
		idx, err := r.Storer.Index()
		if err != nil {
			return err
		}

		for i, entry := range idx.Entries {
			if entry.Name == path {
				idx.Entries = append(idx.Entries[:i], idx.Entries[i+1:]...)
				break
			}
		}

		// write the updated index back to storage
		return r.Storer.SetIndex(idx)
	}

	// HEAD exists, so reset the file to unstage it
	err = w.Reset(&git.ResetOptions{
		Mode:  git.MixedReset,
		Files: []string{path},
	})
	return err
}

// creates a commit with the given message
func commit(message string) error {
	_, w, err := openRepo()
	if err != nil {
		return err
	}

	_, err = w.Commit(message, &git.CommitOptions{})
	return err
}

// returns the name of the current branch
func getCurrentBranch() (string, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return "", err
	}

	head, err := r.Head()
	if err != nil {
		return "", err
	}

	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}

	return "HEAD", nil
}

// returns a list of all local branches
func listBranches() ([]string, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return nil, err
	}

	branches, err := r.Branches()
	if err != nil {
		return nil, err
	}

	var branchNames []string
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		if !strings.Contains(ref.Name().String(), "refs/remotes/") {
			branchNames = append(branchNames, ref.Name().Short())
		}
		return nil
	})

	return branchNames, err
}

// creates a new branch from the current HEAD
func createBranch(branchName string) error {
	r, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	head, err := r.Head()
	if err != nil {
		return err
	}

	// create the new branch
	refName := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(refName, head.Hash())

	return r.Storer.SetReference(ref)
}

// switches to the specified branch
func switchBranch(branchName string) error {
	r, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	return err
}
