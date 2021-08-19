package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/sinmetal/golddog/backend"
	metadatabox "github.com/sinmetalcraft/gcpbox/metadata"
	"go.mercari.io/datastore/clouddatastore"
)

func main() {
	ctx := context.Background()

	pID, err := metadatabox.ProjectID()
	if err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	log.Printf("Listening on port %s", port)

	ds, err := datastore.NewClient(ctx, pID)
	if err != nil {
		panic(err)
	}
	cds, err := clouddatastore.FromClient(ctx, ds)
	if err != nil {
		panic(err)
	}
	gitHubNotifyStore, err := backend.NewGitHubNotifyStore(ctx, cds)
	if err != nil {
		panic(err)
	}
	handlers, err := backend.NewHandlers(ctx, gitHubNotifyStore)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/cron/notifications", handlers.CronNotificationsHandler)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
