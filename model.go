package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	files            []FileStatus
	cursor           int
	quitting         bool
	showInitMenu     bool
	showGitHubAuth   bool
	creatingRepo     bool
	initingRepo      bool
	currentBranch    string
	showBranchMenu   bool
	creatingBranch   bool
	switchingBranch  bool
	showBranchList   bool
	branches         []string
	branchListCursor int
}

type FileStatus struct {
	Path     string
	Status   string // "staged", "unstaged", "untracked"
	Selected bool
}

func NewModel() Model {
	if !isGitRepo() {
		return Model{
			files:            []FileStatus{},
			cursor:           0,
			quitting:         false,
			showInitMenu:     true,
			currentBranch:    "",
			showBranchMenu:   false,
			creatingBranch:   false,
			switchingBranch:  false,
			showBranchList:   false,
			branches:         []string{},
			branchListCursor: 0,
		}
	}

	files, err := getGitStatus()
	if err != nil {
		files = []FileStatus{}
	}

	currentBranch, err := getCurrentBranch()
	if err != nil {
		currentBranch = "unknown"
	}

	return Model{
		files:            files,
		cursor:           0,
		quitting:         false,
		currentBranch:    currentBranch,
		showBranchMenu:   false,
		creatingBranch:   false,
		switchingBranch:  false,
		showBranchList:   false,
		branches:         []string{},
		branchListCursor: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
