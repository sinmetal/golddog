package backend

import "context"

type Handlers struct {
	gitHubNotifyStore *GitHubNotifyStore
}

func NewHandlers(ctx context.Context, gitHubNotifyStore *GitHubNotifyStore) (*Handlers, error) {
	return &Handlers{
		gitHubNotifyStore: gitHubNotifyStore,
	}, nil
}
