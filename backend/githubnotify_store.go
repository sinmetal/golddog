package backend

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.mercari.io/datastore"
	"time"
)

// GitHubNotifyStore is GitHubNotify Entity Store
type GitHubNotifyStore struct {
	ds datastore.Client
}

//GitHubNotifyEntityKind is Kind Name
const GitHubNotifyEntityKind = "GitHubNotify"

// GitHubNotifyEntity is GitHubNotifyをDatastoreに保存するためのEntity
type GitHubNotifyEntity struct {
	ID               string    `json:"id" datastore:"-"`
	Title            string    `json:"title"`
	URL              string    `json:"url"`
	LatestCommentURL string    `json:"latest_comment_url"`
	Type             string    `json:"type"`
	NotifyCount      int       `json:"notifyCount"`
	CreatedAt        time.Time `json:"createdAt"`
}

// Key is Create Key
func (store *GitHubNotifyStore) Key(gitHubNotifyID string) datastore.Key {
	return store.ds.NameKey(GitHubNotifyEntityKind, gitHubNotifyID, nil)
}

// Put to Datastore
func (store *GitHubNotifyStore) Put(ctx context.Context, n *GitHubNotifyEntity) (*GitHubNotifyEntity, error) {
	k := store.ds.NameKey(GitHubNotifyEntityKind, n.ID, nil)
	_, err := store.ds.Put(ctx, k, n)
	if err != nil {
		return nil, err
	}
	return n, nil
}

// Get from Datastore
func (store *GitHubNotifyStore) Get(ctx context.Context, key datastore.Key) (*GitHubNotifyEntity, error) {
	var e GitHubNotifyEntity
	if err := store.ds.Get(ctx, key, &e); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed datastore.Get; key=%+v", key))
	}
	return &e, nil
}
