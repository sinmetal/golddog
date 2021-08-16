package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sinmetal/golddog/backend"
	"google.golang.org/appengine"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	http.HandleFunc("/cron/notifications", backend.CronNotificationsHandler)

	appengine.Main()
}
