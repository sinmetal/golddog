package backend

import (
	"context"

	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
)

type GitHubClient struct {
	c *github.Client
}

func NewGitHubClient(ctx context.Context, token string) *GitHubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return &GitHubClient{c: client}
}

func (c *GitHubClient) ListNotifications(ctx context.Context) ([]*github.Notification, error) {
	l, _, err := c.c.Activity.ListNotifications(ctx, &github.NotificationListOptions{Participating: true})
	return l, err
}
