package main

import (
	"errors"
	"fmt"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
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

type commitFormCompleteMsg struct {
	message string
	err     error
}

type createBranchFormCompleteMsg struct {
	branchName string
	err        error
}

type switchBranchFormCompleteMsg struct {
	branchName string
	err        error
}

type initLocalRepoFormCompleteMsg struct {
	defaultBranch string
	err           error
}

type githubSetupFormCompleteMsg struct {
	token         string
	repoName      string
	repoDesc      string
	repoPrivate   bool
	defaultBranch string
	err           error
}

type formExecCommand struct {
	run func() error
}

func (c formExecCommand) Run() error {
	return c.run()
}

func (formExecCommand) SetStdin(io.Reader)  {}
func (formExecCommand) SetStdout(io.Writer) {}
func (formExecCommand) SetStderr(io.Writer) {}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			if m.showInitMenu {
				m.showInitMenu = false
				return m, runInitLocalRepoFormCmd()
			}
			if m.showBranchMenu {
				m.showBranchMenu = false
				return m, runCreateBranchFormCmd()
			}
		case "2":
			if m.showInitMenu {
				m.showInitMenu = false
				m.showGitHubAuth = true
				return m, nil
			}
			if m.showBranchMenu {
				m.showBranchMenu = false
				return m, runSwitchBranchFormCmd()
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
				m.showGitHubAuth = false
				return m, runGitHubSetupFormCmd()
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
			return m, runCommitFormCmd()

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
	case commitFormCompleteMsg:
		if msg.err != nil {
			if errors.Is(msg.err, huh.ErrUserAborted) {
				return m, showToast("commit cancelled", ToastInfo)
			}
			return m, showToast(fmt.Sprintf("failed to open commit form: %v", msg.err), ToastError)
		}
		if msg.message == "" {
			return m, showToast("commit cancelled", ToastInfo)
		}
		return m, m.commitWithMessage(msg.message)
	case createBranchFormCompleteMsg:
		if msg.err != nil {
			if errors.Is(msg.err, huh.ErrUserAborted) {
				return m, showToast("branch creation cancelled", ToastInfo)
			}
			return m, showToast(fmt.Sprintf("failed to open branch form: %v", msg.err), ToastError)
		}
		if msg.branchName == "" {
			return m, showToast("branch creation cancelled", ToastInfo)
		}
		m.creatingBranch = true
		return m, m.createBranch(msg.branchName)
	case switchBranchFormCompleteMsg:
		if msg.err != nil {
			if errors.Is(msg.err, huh.ErrUserAborted) {
				return m, showToast("branch switch cancelled", ToastInfo)
			}
			return m, showToast(fmt.Sprintf("failed to open branch selector: %v", msg.err), ToastError)
		}
		if msg.branchName == "" {
			return m, showToast("branch switch cancelled", ToastInfo)
		}
		m.switchingBranch = true
		return m, m.switchBranch(msg.branchName)
	case initLocalRepoFormCompleteMsg:
		if msg.err != nil {
			m.showInitMenu = true
			if errors.Is(msg.err, huh.ErrUserAborted) {
				return m, showToast("repository initialization cancelled", ToastInfo)
			}
			return m, showToast(fmt.Sprintf("failed to open initialization form: %v", msg.err), ToastError)
		}
		if msg.defaultBranch == "" {
			m.showInitMenu = true
			return m, showToast("repository initialization cancelled", ToastInfo)
		}
		m.initingRepo = true
		return m, m.initLocalRepo(msg.defaultBranch)
	case githubSetupFormCompleteMsg:
		if msg.err != nil {
			m.showInitMenu = true
			if errors.Is(msg.err, huh.ErrUserAborted) {
				return m, showToast("github setup cancelled", ToastInfo)
			}
			return m, showToast(fmt.Sprintf("GitHub token error: %v", msg.err), ToastError)
		}
		if msg.token == "" || msg.repoName == "" {
			m.showInitMenu = true
			return m, showToast("github setup cancelled", ToastInfo)
		}
		m.creatingRepo = true
		return m, m.createGitHubRepo(msg.token, msg.repoName, msg.repoDesc, msg.repoPrivate, msg.defaultBranch)
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

func runCommitFormCmd() tea.Cmd {
	var message string
	return tea.Exec(formExecCommand{
		run: func() error {
			var err error
			message, err = showCommitForm()
			return err
		},
	}, func(err error) tea.Msg {
		return commitFormCompleteMsg{message: message, err: err}
	})
}

func runCreateBranchFormCmd() tea.Cmd {
	var branchName string
	return tea.Exec(formExecCommand{
		run: func() error {
			var err error
			branchName, err = showCreateBranchForm()
			return err
		},
	}, func(err error) tea.Msg {
		return createBranchFormCompleteMsg{branchName: branchName, err: err}
	})
}

func runSwitchBranchFormCmd() tea.Cmd {
	var branchName string
	return tea.Exec(formExecCommand{
		run: func() error {
			branches, err := listBranches()
			if err != nil {
				return err
			}
			branchName, err = showBranchSelectionForm(branches)
			return err
		},
	}, func(err error) tea.Msg {
		return switchBranchFormCompleteMsg{branchName: branchName, err: err}
	})
}

func runInitLocalRepoFormCmd() tea.Cmd {
	var defaultBranch string
	return tea.Exec(formExecCommand{
		run: func() error {
			var err error
			defaultBranch, err = showLocalRepoForm()
			return err
		},
	}, func(err error) tea.Msg {
		return initLocalRepoFormCompleteMsg{defaultBranch: defaultBranch, err: err}
	})
}

func runGitHubSetupFormCmd() tea.Cmd {
	var (
		token         string
		repoName      string
		repoDesc      string
		repoPrivate   bool
		defaultBranch string
	)

	return tea.Exec(formExecCommand{
		run: func() error {
			var err error
			token, err = getGitHubToken()
			if err != nil {
				return err
			}
			repoName, repoDesc, repoPrivate, defaultBranch, err = showGitHubRepoForm(token)
			return err
		},
	}, func(err error) tea.Msg {
		return githubSetupFormCompleteMsg{
			token:         token,
			repoName:      repoName,
			repoDesc:      repoDesc,
			repoPrivate:   repoPrivate,
			defaultBranch: defaultBranch,
			err:           err,
		}
	})
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
func (m *Model) commitWithMessage(message string) tea.Cmd {
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

type commitSuccessMsg struct{}

func (m *Model) handleCommitSuccess() tea.Cmd {
	m.refreshFiles()
	return showToast("changes committed successfully", ToastSuccess)
}

// shows create branch form and creates branch
func (m *Model) createBranch(branchName string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		err := createBranch(branchName)
		if err != nil {
			return createBranchErrorMsg{err: err}
		}
		return createBranchCompleteMsg{branchName: branchName}
	})
}

// shows switch branch form and switches branch
func (m *Model) switchBranch(branchName string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		err := switchBranch(branchName)
		if err != nil {
			return switchBranchErrorMsg{err: err}
		}
		return switchBranchCompleteMsg{branchName: branchName}
	})
}

// shows local repo form and initializes repository
func (m *Model) initLocalRepo(defaultBranch string) tea.Cmd {
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

func (m *Model) createGitHubRepo(token, repoName, repoDesc string, repoPrivate bool, defaultBranch string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
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
