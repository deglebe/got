package main

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/google/go-github/v61/github"
	"golang.org/x/oauth2"
)

// creates a github client with the provided token
func createGitHubClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// creates a new repository on github and sets up remote origin
func createGitHubRepo(token, name, description string, private bool, defaultBranch string) (*github.Repository, error) {
	client := createGitHubClient(token)
	ctx := context.Background()

	repo := &github.Repository{
		Name:          github.String(name),
		Description:   github.String(description),
		Private:       github.Bool(private),
		DefaultBranch: github.String(defaultBranch),
	}

	createdRepo, _, err := client.Repositories.Create(ctx, "", repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create github repository: %w", err)
	}

	// add remote origin to local repository
	if err := addRemoteOrigin(createdRepo.GetCloneURL()); err != nil {
		return nil, fmt.Errorf("failed to add remote origin: %w", err)
	}

	// rename default branch to match user preference
	if defaultBranch != "" {
		cmd := exec.Command("git", "branch", "-m", defaultBranch)
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to rename branch: %w", err)
		}
	}

	return createdRepo, nil
}

// validates if a github token is working
func validateGitHubToken(token string) error {
	client := createGitHubClient(token)
	ctx := context.Background()

	_, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("invalid github token: %w", err)
	}

	return nil
}
