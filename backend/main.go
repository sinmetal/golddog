package backend

import (
	"context"

	"go.mercari.io/datastore"
	"go.mercari.io/datastore/aedatastore"
)

// FromContext is Create Datastore Client from Context
func FromContext(ctx context.Context) (datastore.Client, error) {
	return aedatastore.FromContext(ctx)
}
