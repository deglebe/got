package main

import (
	"fmt"
	"time"

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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			if m.showInitMenu {
				m.showInitMenu = false
				return m, m.initLocalRepo()
			}
			if m.showBranchMenu {
				m.showBranchMenu = false
				return m, m.createBranchForm()
			}
		case "2":
			if m.showInitMenu {
				m.showInitMenu = false
				m.showGitHubAuth = true
				return m, nil
			}
			if m.showBranchMenu {
				m.showBranchMenu = false
				return m, m.switchBranchForm()
			}
		case "enter":
			if m.showBranchList && m.branchListCursor < len(m.branches) {
				selectedBranch := m.branches[m.branchListCursor]
				if selectedBranch != m.currentBranch {
					m.showBranchList = false
					m.switchingBranch = true
					return m, tea.Cmd(func() tea.Msg {
						err := switchBranch(selectedBranch)
						if err != nil {
							return switchBranchErrorMsg{err: err}
						}
						return switchBranchCompleteMsg{branchName: selectedBranch}
					})
				}
				return m, nil
			}
			if m.showGitHubAuth {
				// get github token
				token, err := getGitHubToken()
				if err != nil {
					m.showGitHubAuth = false
					m.showInitMenu = true
					return m, showToast(fmt.Sprintf("GitHub token error: %v", err), ToastError)
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
			if m.showBranchList {
				m.showBranchList = false
				return m, nil
			}
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.showBranchList {
				if m.branchListCursor > 0 {
					m.branchListCursor--
				}
			} else if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.showBranchList {
				if m.branchListCursor < len(m.branches)-1 {
					m.branchListCursor++
				}
			} else if m.cursor < len(m.files)-1 {
				m.cursor++
			}

		case " ":
			if m.cursor < len(m.files) {
				m.files[m.cursor].Selected = !m.files[m.cursor].Selected
			}

		case "s":
			m.stageSelectedFiles()
			m.refreshFiles()
			if m.showToast {
				return m, dismissToastAfter(3 * time.Second)
			}

		case "u":
			m.unstageSelectedFiles()
			m.refreshFiles()
			if m.showToast {
				return m, dismissToastAfter(3 * time.Second)
			}

		case "c":
			return m, m.commitChanges()

		case "b":
			if !m.showInitMenu && !m.showGitHubAuth && !m.creatingRepo && !m.initingRepo {
				m.showBranchMenu = true
				return m, nil
			}

		case "3":
			if m.showBranchMenu {
				m.showBranchMenu = false
				m.showBranchList = true
				branches, err := listBranches()
				if err != nil {
					branches = []string{}
					return m, showToast(fmt.Sprintf("failed to list branches: %v", err), ToastError)
				}
				m.branches = branches
				m.branchListCursor = 0
				return m, nil
			}
		}
	}

	switch msg := msg.(type) {
	case showToastMsg:
		m.toast = msg.toast
		m.showToast = true
		return m, dismissToastAfter(3 * time.Second)
	case dismissToastMsg:
		m.showToast = false
		m.toast = Toast{}
		return m, nil
	case initCompleteMsg:
		m.initingRepo = false
		m.files = msg.files
		currentBranch, err := getCurrentBranch()
		if err == nil {
			m.currentBranch = currentBranch
		}
		return m, nil
	case initErrorMsg:
		m.initingRepo = false
		m.showInitMenu = true
		return m, showToast(fmt.Sprintf("failed to initialize repository: %v", msg.err), ToastError)
	case githubRepoCompleteMsg:
		m.creatingRepo = false
		// load files and show main
		files, err := getGitStatus()
		if err != nil {
			files = []FileStatus{}
		}
		m.files = files
		// set current branch after repo creation
		currentBranch, err := getCurrentBranch()
		if err == nil {
			m.currentBranch = currentBranch
		}
		return m, showToast(fmt.Sprintf("repository created successfully: %s", msg.repoURL), ToastSuccess)
	case githubRepoErrorMsg:
		m.creatingRepo = false
		m.showGitHubAuth = true
		return m, showToast(fmt.Sprintf("failed to create github repository: %v", msg.err), ToastError)
	case createBranchCompleteMsg:
		m.creatingBranch = false
		// update current branch and refresh files
		currentBranch, err := getCurrentBranch()
		if err == nil {
			m.currentBranch = currentBranch
		}
		m.refreshFiles()
		return m, showToast(fmt.Sprintf("branch '%s' created successfully", msg.branchName), ToastSuccess)
	case createBranchErrorMsg:
		m.creatingBranch = false
		return m, showToast(fmt.Sprintf("failed to create branch: %v", msg.err), ToastError)
	case switchBranchCompleteMsg:
		m.switchingBranch = false
		if currentBranch, err := getCurrentBranch(); err == nil {
			m.currentBranch = currentBranch
		} else {
			m.currentBranch = msg.branchName
		}
		m.refreshFiles()
		return m, showToast(fmt.Sprintf("switched to branch '%s'", msg.branchName), ToastSuccess)
	case switchBranchErrorMsg:
		m.switchingBranch = false
		return m, showToast(fmt.Sprintf("failed to switch branch: %v", msg.err), ToastError)
	case commitSuccessMsg:
		return m, m.handleCommitSuccess()
	}

	return m, nil
}

// stages all selected files
func (m *Model) stageSelectedFiles() {
	var stagedCount int
	var errors []error
	for _, file := range m.files {
		if file.Selected && file.Status != "staged" {
			if err := stageFile(file.Path); err != nil {
				errors = append(errors, fmt.Errorf("%s: %w", file.Path, err))
			} else {
				stagedCount++
			}
		}
	}
	if len(errors) > 0 {
		// show error for first failed file
		m.showToast = true
		m.toast = Toast{
			Message: fmt.Sprintf("failed to stage some files: %v", errors[0]),
			Type:    ToastError,
		}
	} else if stagedCount > 0 {
		m.showToast = true
		m.toast = Toast{
			Message: fmt.Sprintf("staged %d file(s)", stagedCount),
			Type:    ToastSuccess,
		}
	}
}

// unstages all selected files
func (m *Model) unstageSelectedFiles() {
	var unstagedCount int
	var errors []error
	for _, file := range m.files {
		if file.Selected && file.Status == "staged" {
			if err := unstageFile(file.Path); err != nil {
				errors = append(errors, fmt.Errorf("%s: %w", file.Path, err))
			} else {
				unstagedCount++
			}
		}
	}
	if len(errors) > 0 {
		// show error for first failed file
		m.showToast = true
		m.toast = Toast{
			Message: fmt.Sprintf("failed to unstage some files: %v", errors[0]),
			Type:    ToastError,
		}
	} else if unstagedCount > 0 {
		m.showToast = true
		m.toast = Toast{
			Message: fmt.Sprintf("unstaged %d file(s)", unstagedCount),
			Type:    ToastSuccess,
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
func (m *Model) commitChanges() tea.Cmd {
	if message, err := showCommitForm(); err == nil && message != "" {
		return tea.Cmd(func() tea.Msg {
			if err := commit(message); err != nil {
				return showToastMsg{
					toast: Toast{
						Message: fmt.Sprintf("failed to commit: %v", err),
						Type:    ToastError,
					},
				}
			}
			return commitSuccessMsg{}
		})
	}
	return nil
}

type commitSuccessMsg struct{}

func (m *Model) handleCommitSuccess() tea.Cmd {
	m.refreshFiles()
	return showToast("changes committed successfully", ToastSuccess)
}

// shows create branch form and creates branch
func (m *Model) createBranchForm() tea.Cmd {
	branchName, err := showCreateBranchForm()
	if err != nil {
		return showToast("branch creation cancelled", ToastInfo)
	}
	if branchName != "" {
		return tea.Cmd(func() tea.Msg {
			err := createBranch(branchName)
			if err != nil {
				return createBranchErrorMsg{err: err}
			}
			return createBranchCompleteMsg{branchName: branchName}
		})
	}
	return nil
}

// shows switch branch form and switches branch
func (m *Model) switchBranchForm() tea.Cmd {
	branches, err := listBranches()
	if err != nil {
		return showToast(fmt.Sprintf("failed to list branches: %v", err), ToastError)
	}
	branchName, err := showBranchSelectionForm(branches)
	if err != nil {
		return showToast("branch switch cancelled", ToastInfo)
	}
	if branchName != "" {
		return tea.Cmd(func() tea.Msg {
			err := switchBranch(branchName)
			if err != nil {
				return switchBranchErrorMsg{err: err}
			}
			return switchBranchCompleteMsg{branchName: branchName}
		})
	}
	return nil
}

// shows local repo form and initializes repository
func (m *Model) initLocalRepo() tea.Cmd {
	defaultBranch, err := showLocalRepoForm()
	if err != nil {
		return showToast("repository initialization cancelled", ToastInfo)
	}
	if defaultBranch != "" {
		return tea.Cmd(func() tea.Msg {
			err := initGitRepoWithBranch(defaultBranch)
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
	return nil
}
