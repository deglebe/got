package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	files          []FileStatus
	cursor         int
	quitting       bool
	showInitMenu   bool
	showGitHubAuth bool
	creatingRepo   bool
	initingRepo    bool
}

type FileStatus struct {
	Path     string
	Status   string // "staged", "unstaged", "untracked"
	Selected bool
}

func NewModel() Model {
	if !isGitRepo() {
		return Model{
			files:        []FileStatus{},
			cursor:       0,
			quitting:     false,
			showInitMenu: true,
		}
	}

	files, err := getGitStatus()
	if err != nil {
		files = []FileStatus{}
	}

	return Model{
		files:    files,
		cursor:   0,
		quitting: false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
