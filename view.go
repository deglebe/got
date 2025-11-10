package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))

	statusStyle = func(status string) lipgloss.Style {
		switch status {
		case "staged":
			return lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
		case "unstaged":
			return lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
		case "untracked":
			return lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow
		default:
			return lipgloss.NewStyle()
		}
	}

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			MarginTop(1)
)

// initial menu for repo setup
func (m Model) renderInitMenu() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("got"))
	b.WriteString("\n\n")
	b.WriteString("no git repository found in current directory.\n\n")
	b.WriteString("choose an option:\n\n")
	b.WriteString("1. initialize local git repository\n")
	b.WriteString("2. create github repository & initialize locally\n")
	b.WriteString("q. quit\n\n")
	b.WriteString(helpStyle.Render("press 1, 2, or q"))

	return b.String()
}

// github auth form
func (m Model) renderGitHubAuth() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("github repository setup"))
	b.WriteString("\n\n")
	b.WriteString("this will create a new github repository and initialize it locally\n\n")
	b.WriteString("if you haven't set a github access token, enter one below\n")
	b.WriteString("the token will be saved to ~/.config/got/config.yaml for future use\n\n")
	b.WriteString("create a token at: https://github.com/settings/tokens\n")
	b.WriteString("required scopes: repo, workflow\n\n")
	b.WriteString(helpStyle.Render("press enter to continue, esc to go back"))

	return b.String()
}

// branch menu for branch operations
func (m Model) renderBranchMenu() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("branch management"))
	b.WriteString("\n\n")
	b.WriteString("current branch: " + m.currentBranch + "\n\n")
	b.WriteString("choose an option:\n\n")
	b.WriteString("1. create new branch\n")
	b.WriteString("2. switch branch\n")
	b.WriteString("3. list all branches\n")
	b.WriteString("esc. back to main menu\n\n")
	b.WriteString(helpStyle.Render("press 1, 2, 3, or esc"))

	return b.String()
}

// branch list view for viewing all branches
func (m Model) renderBranchList() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("all branches"))
	b.WriteString("\n\n")
	b.WriteString("current branch: " + m.currentBranch + "\n\n")

	if len(m.branches) == 0 {
		b.WriteString("no branches found.\n")
	} else {
		for i, branch := range m.branches {
			cursor := " "
			if m.branchListCursor == i {
				cursor = cursorStyle.Render(">")
			}

			branchDisplay := branch
			if branch == m.currentBranch {
				branchDisplay = selectedStyle.Render("* " + branch)
			}

			line := fmt.Sprintf("%s %s", cursor, branchDisplay)
			if m.branchListCursor == i {
				line = cursorStyle.Render(line)
			}

			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ or j/k: navigate • enter: switch to branch • esc: back"))

	return b.String()
}

func (m Model) View() string {
	if m.quitting {
		return "goodbye!\n"
	}

	if m.showInitMenu {
		return m.renderInitMenu()
	}

	if m.showGitHubAuth {
		return m.renderGitHubAuth()
	}

	if m.creatingRepo {
		return "creating github repository...\n"
	}

	if m.initingRepo {
		return "initializing git repository...\n"
	}

	if m.showBranchMenu {
		return m.renderBranchMenu()
	}

	if m.creatingBranch {
		return "creating branch...\n"
	}

	if m.switchingBranch {
		return "switching branch...\n"
	}

	if m.showBranchList {
		return m.renderBranchList()
	}

	var b strings.Builder

	title := "got"
	if m.currentBranch != "" {
		title += " (" + m.currentBranch + ")"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	if len(m.files) == 0 {
		b.WriteString("no changes to stage.\n")
	} else {
		for i, file := range m.files {
			cursor := " "
			if m.cursor == i {
				cursor = cursorStyle.Render(">")
			}

			checkbox := "[ ]"
			if file.Selected {
				checkbox = selectedStyle.Render("[x]")
			}

			status := statusStyle(file.Status).Render(fmt.Sprintf("[%s]", file.Status))
			line := fmt.Sprintf("%s %s %s %s", cursor, checkbox, status, file.Path)

			if m.cursor == i {
				line = cursorStyle.Render(line)
			}

			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if len(m.files) == 0 {
		b.WriteString(helpStyle.Render("b: branches • c: commit • q: quit"))
	} else {
		b.WriteString(helpStyle.Render("↑/↓ or j/k: navigate • space: toggle selection • s: stage selected • u: unstage selected • b: branches • c: commit • q: quit"))
	}

	return b.String()
}
