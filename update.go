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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "2":
			if m.showInitMenu {
				m.showInitMenu = false
				m.showGitHubAuth = true
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
		}
	}

	switch msg := msg.(type) {
	case initCompleteMsg:
		m.initingRepo = false
		m.files = msg.files
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
		// TODO: show success message with repo url
		return m, nil
	case githubRepoErrorMsg:
		m.creatingRepo = false
		m.showGitHubAuth = true
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
