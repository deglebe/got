package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type initCompleteMsg struct {
	files []FileStatus
}

type initErrorMsg struct {
	err error
}

type githubRepoCompleteMsg struct {
	repoURL string
}

type githubRepoErrorMsg struct {
	err error
}

type createBranchCompleteMsg struct {
	branchName string
}

type createBranchErrorMsg struct {
	err error
}

type switchBranchCompleteMsg struct {
	branchName string
}

type switchBranchErrorMsg struct {
	err error
}

type listBranchesCompleteMsg struct {
	branches []string
}

type createBranchFormCompleteMsg struct {
	branchName string
}

type createBranchFormErrorMsg struct {
	err error
}

type switchBranchFormCompleteMsg struct {
	branchName string
}

type switchBranchFormErrorMsg struct {
	err error
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle forms synchronously when they are active
	if m.showCreateBranchForm {
		branchName, err := showCreateBranchForm()
		if err != nil {
			m.showCreateBranchForm = false
			return m, tea.Cmd(func() tea.Msg {
				return createBranchFormErrorMsg{err: err}
			})
		}
		m.showCreateBranchForm = false
		return m, tea.Cmd(func() tea.Msg {
			err := createBranch(branchName)
			if err != nil {
				return createBranchErrorMsg{err: err}
			}
			return createBranchCompleteMsg{branchName: branchName}
		})
	}

	if m.showSwitchBranchForm {
		branches, err := listBranches()
		if err != nil {
			m.showSwitchBranchForm = false
			return m, tea.Cmd(func() tea.Msg {
				return switchBranchFormErrorMsg{err: err}
			})
		}
		branchName, err := showBranchSelectionForm(branches)
		if err != nil {
			m.showSwitchBranchForm = false
			return m, tea.Cmd(func() tea.Msg {
				return switchBranchFormErrorMsg{err: err}
			})
		}
		m.showSwitchBranchForm = false
		return m, tea.Cmd(func() tea.Msg {
			err := switchBranch(branchName)
			if err != nil {
				return switchBranchErrorMsg{err: err}
			}
			return switchBranchCompleteMsg{branchName: branchName}
		})
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			if m.showInitMenu {
				m.initingRepo = true
				m.showInitMenu = false
				return m, tea.Cmd(func() tea.Msg {
					err := initGitRepo()
					if err != nil {
						return initErrorMsg{err: err}
					}
					files, err := getGitStatus()
					if err != nil {
						files = []FileStatus{}
					}
					return initCompleteMsg{files: files}
				})
			}
			if m.showBranchMenu {
				m.showBranchMenu = false
				m.showCreateBranchForm = true
				return m, nil
			}
		case "2":
			if m.showInitMenu {
				m.showInitMenu = false
				m.showGitHubAuth = true
				return m, nil
			}
			if m.showBranchMenu {
				m.showBranchMenu = false
				m.showSwitchBranchForm = true
				return m, nil
			}
		case "enter":
			if m.showGitHubAuth {
				// get github token
				token, err := getGitHubToken()
				if err != nil {
					m.showGitHubAuth = false
					m.showInitMenu = true
					return m, nil
				}

				// token valid: show repo form
				repoName, repoDesc, repoPrivate, defaultBranch, err := showGitHubRepoForm(token)
				if err != nil {
					return m, nil
				}

				m.creatingRepo = true
				m.showGitHubAuth = false

				return m, tea.Cmd(func() tea.Msg {
					if err := initGitRepo(); err != nil {
						return githubRepoErrorMsg{err: err}
					}

					repo, err := createGitHubRepo(token, repoName, repoDesc, repoPrivate, defaultBranch)
					if err != nil {
						return githubRepoErrorMsg{err: err}
					}

					return githubRepoCompleteMsg{repoURL: repo.GetHTMLURL()}
				})
			}
		case "esc":
			if m.showGitHubAuth {
				m.showGitHubAuth = false
				m.showInitMenu = true
				return m, nil
			}
			if m.showBranchMenu {
				m.showBranchMenu = false
				return m, nil
			}
			if m.showCreateBranchForm {
				m.showCreateBranchForm = false
				m.showBranchMenu = true
				return m, nil
			}
			if m.showSwitchBranchForm {
				m.showSwitchBranchForm = false
				m.showBranchMenu = true
				return m, nil
			}
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
			}

		case " ":
			if m.cursor < len(m.files) {
				m.files[m.cursor].Selected = !m.files[m.cursor].Selected
			}

		case "s":
			m.stageSelectedFiles()
			m.refreshFiles()

		case "u":
			m.unstageSelectedFiles()
			m.refreshFiles()

		case "c":
			m.commitChanges()

		case "b":
			if !m.showInitMenu && !m.showGitHubAuth && !m.creatingRepo && !m.initingRepo {
				m.showBranchMenu = true
				return m, nil
			}

		case "3":
			if m.showBranchMenu {
				return m, tea.Cmd(func() tea.Msg {
					branches, err := listBranches()
					if err != nil {
						return listBranchesCompleteMsg{branches: []string{}}
					}
					return listBranchesCompleteMsg{branches: branches}
				})
			}
		}
	}

	switch msg := msg.(type) {
	case initCompleteMsg:
		m.initingRepo = false
		m.files = msg.files
		// Set current branch after repo initialization
		currentBranch, err := getCurrentBranch()
		if err == nil {
			m.currentBranch = currentBranch
		}
		return m, nil
	case initErrorMsg:
		m.initingRepo = false
		m.showInitMenu = true
		// TODO: show error message
		return m, nil
	case githubRepoCompleteMsg:
		m.creatingRepo = false
		// load files and show main
		files, err := getGitStatus()
		if err != nil {
			files = []FileStatus{}
		}
		m.files = files
		// Set current branch after repo creation
		currentBranch, err := getCurrentBranch()
		if err == nil {
			m.currentBranch = currentBranch
		}
		// TODO: show success message with repo url
		return m, nil
	case githubRepoErrorMsg:
		m.creatingRepo = false
		m.showGitHubAuth = true
		// TODO: show error message
		return m, nil
	case createBranchCompleteMsg:
		m.creatingBranch = false
		// Update current branch and refresh files
		currentBranch, err := getCurrentBranch()
		if err == nil {
			m.currentBranch = currentBranch
		}
		m.refreshFiles()
		return m, nil
	case createBranchErrorMsg:
		m.creatingBranch = false
		m.showBranchMenu = true
		// TODO: show error message
		return m, nil
	case switchBranchCompleteMsg:
		m.switchingBranch = false
		m.currentBranch = msg.branchName
		m.refreshFiles()
		return m, nil
	case switchBranchErrorMsg:
		m.switchingBranch = false
		m.showBranchMenu = true
		// TODO: show error message
		return m, nil
	case listBranchesCompleteMsg:
		// For now, just show the branches in the menu - could be enhanced to show in a separate view
		// TODO: implement branch listing view
		return m, nil
	case createBranchFormCompleteMsg:
		m.creatingBranch = false
		// Update current branch and refresh files
		currentBranch, err := getCurrentBranch()
		if err == nil {
			m.currentBranch = currentBranch
		}
		m.refreshFiles()
		return m, nil
	case createBranchFormErrorMsg:
		m.creatingBranch = false
		m.showBranchMenu = true
		// TODO: show error message
		return m, nil
	case switchBranchFormCompleteMsg:
		m.switchingBranch = false
		m.currentBranch = msg.branchName
		m.refreshFiles()
		return m, nil
	case switchBranchFormErrorMsg:
		m.switchingBranch = false
		m.showBranchMenu = true
		// TODO: show error message
		return m, nil
	}

	return m, nil
}

// stages all selected files
func (m *Model) stageSelectedFiles() {
	for _, file := range m.files {
		if file.Selected && file.Status != "staged" {
			stageFile(file.Path)
		}
	}
}

// unstages all selected files
func (m *Model) unstageSelectedFiles() {
	for _, file := range m.files {
		if file.Selected && file.Status == "staged" {
			unstageFile(file.Path)
		}
	}
}

// reloads the file list and adjusts cursor position
func (m *Model) refreshFiles() {
	if files, err := getGitStatus(); err == nil {
		m.files = files
		// keep cursor in bounds
		if m.cursor >= len(m.files) {
			m.cursor = len(m.files) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
	}
}

// shows commit form and commits changes
func (m *Model) commitChanges() {
	if message, err := showCommitForm(); err == nil && message != "" {
		if err := commit(message); err == nil {
			m.refreshFiles()
		}
	}
}
