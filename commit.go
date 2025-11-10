package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// display a huh form for github token input
func showGitHubTokenForm() (string, error) {
	var token string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("github personal access token").
				Description("enter your token (will be hidden, esc to cancel)").
				Placeholder("ghp_...").
				Value(&token).
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("token cannot be empty")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCharm())

	err := form.Run()
	return token, err
}

// display a huh form for GitHub repository creation
func showGitHubRepoForm(token string) (string, string, bool, string, error) {
	var repoName string
	var repoDesc string
	var repoPrivate bool
	var defaultBranch string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("repository name").
				Description("choose a name for your repository (esc to cancel)").
				Placeholder("my_project").
				Value(&repoName).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("repository name cannot be empty")
					}
					return nil
				}),
			huh.NewInput().
				Title("description (optional)").
				Description("add a description for your repository").
				Placeholder("the greatest project known to humanity").
				Value(&repoDesc),
			huh.NewInput().
				Title("default branch").
				Description("default branch name for the repository").
				Placeholder("main").
				Value(&defaultBranch).
				Validate(func(s string) error {
					if s == "" {
						defaultBranch = "main" // set default if empty
						return nil
					}
					return nil
				}),
			huh.NewConfirm().
				Title("private Repository?").
				Description("make this repository private").
				Value(&repoPrivate),
		),
	).WithTheme(huh.ThemeCharm())

	err := form.Run()
	return repoName, repoDesc, repoPrivate, defaultBranch, err
}

// display a huh form for comprehensive commit message input
func showCommitForm() (string, error) {
	var commitType string
	var commitScope string
	var commitSubject string
	var commitBody string

	// predefined commit types
	types := []huh.Option[string]{
		huh.NewOption("feat: a new feature", "feat"),
		huh.NewOption("fix: a bug fix", "fix"),
		huh.NewOption("docs: documentation only changes", "docs"),
		huh.NewOption("style: changes that do not affect the meaning of the code", "style"),
		huh.NewOption("refactor: a code change that neither fixes a bug nor adds a feature", "refactor"),
		huh.NewOption("test: adding missing tests or correcting existing tests", "test"),
		huh.NewOption("chore: changes to the build process or auxiliary tools", "chore"),
		huh.NewOption("perf: a code change that improves performance", "perf"),
		huh.NewOption("ci: changes to CI configuration files and scripts", "ci"),
		huh.NewOption("build: changes that affect the build system or external dependencies", "build"),
		huh.NewOption("revert: reverts a previous commit", "revert"),
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("commit type").
				Description("select the type of change you're committing (esc to cancel)").
				Options(types...).
				Value(&commitType),
			huh.NewInput().
				Title("scope (optional)").
				Description("the scope of the change (e.g. component name)").
				Placeholder("auth, api, ui").
				Value(&commitScope),
			huh.NewInput().
				Title("subject").
				Description("brief description of the change").
				Placeholder("add user auth route").
				Value(&commitSubject).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("commit subject cannot be empty")
					}
					if len(s) > 72 {
						return fmt.Errorf("commit subject should be 72 characters or less")
					}
					return nil
				}),
			huh.NewText().
				Title("body (optional)").
				Description("detailed description of the change").
				Placeholder("this change implements user authentication with...").
				Value(&commitBody),
		),
	).WithTheme(huh.ThemeCharm())

	err := form.Run()
	if err != nil {
		return "", err
	}

	var commitMessage string
	if commitScope != "" {
		commitMessage = fmt.Sprintf("%s(%s): %s", commitType, commitScope, commitSubject)
	} else {
		commitMessage = fmt.Sprintf("%s: %s", commitType, commitSubject)
	}

	if commitBody != "" {
		commitMessage += "\n\n" + commitBody
	}

	return commitMessage, nil
}

// display a huh form for branch creation
func showCreateBranchForm() (string, error) {
	var branchName string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("new branch name").
				Description("enter the name for the new branch (esc to cancel)").
				Placeholder("feature/my-feature").
				Value(&branchName).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("branch name cannot be empty")
					}
					if strings.Contains(s, " ") {
						return fmt.Errorf("branch name cannot contain spaces")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCharm())

	err := form.Run()
	return branchName, err
}

// display a huh form for branch selection
func showBranchSelectionForm(branches []string) (string, error) {
	if len(branches) == 0 {
		return "", fmt.Errorf("no branches available")
	}

	var selectedBranch string

	options := make([]huh.Option[string], len(branches))
	for i, branch := range branches {
		options[i] = huh.NewOption(branch, branch)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("select branch").
				Description("choose a branch to switch to (esc to cancel)").
				Options(options...).
				Value(&selectedBranch),
		),
	).WithTheme(huh.ThemeCharm())

	err := form.Run()
	return selectedBranch, err
}
