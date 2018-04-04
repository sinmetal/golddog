package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.mercari.io/datastore"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

func init() {
	http.HandleFunc("/cron/notifications", handler)
}

// GitHubNotification is GitHubのNotificationの構造体
// 必要な項目のみ列挙している
type GitHubNotification struct {
	ID      string                    `json:"id"`
	Subject GitHubNotificationSubject `json:"subject"`
}

// GitHubNotificationSubject is GitHubのNotificationのSubjectの構造体
type GitHubNotificationSubject struct {
	Title            string `json:"title"`
	URL              string `json:"url"`
	LatestCommentURL string `json:"latest_comment_url"`
	Type             string `json:"type"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	ac := GetAppConfig(ctx)

	client := urlfetch.Client(ctx)
	req, err := http.NewRequest("GET", "https://api.github.com/notifications?participating=true", nil)
	if err != nil {
		log.Errorf(ctx, "%+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", ac.GitHubToken))
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(ctx, "%+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf(ctx, "%+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Infof(ctx, "%s", string(body))
	var ns []GitHubNotification
	if err := json.Unmarshal(body, &ns); err != nil {
		log.Errorf(ctx, "%+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(ns) < 1 {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text-plain")
		w.Write([]byte("nothing update"))
		return
	}

	ds, err := FromContext(ctx)
	if err != nil {
		log.Errorf(ctx, "%+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	store := GitHubNotifyStore{
		ds,
	}

	for _, n := range ns {
		key := store.Key(n.ID)
		e, err := store.Get(ctx, key)
		if errors.Cause(err) == datastore.ErrNoSuchEntity {
			e.ID = n.ID
			e.Title = n.Subject.Title
			e.URL = n.Subject.URL
			e.LatestCommentURL = n.Subject.LatestCommentURL
			e.Type = n.Subject.Type
			e.NotifyCount = 0
			e.CreatedAt = time.Now()
		} else if err != nil {
			log.Errorf(ctx, "%+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		t := e.CreatedAt.Add(time.Duration(e.NotifyCount) * time.Minute * 45)
		if e.NotifyCount > 0 || t.After(time.Now()) {
			log.Infof(ctx, "not snooze...")
			continue
		}

		msg := buildMessage(e)
		if err := PostMessage(ctx, msg); err != nil {
			log.Errorf(ctx, "failed slack.post %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		e.NotifyCount++
		_, err = store.Put(ctx, e)
		if err != nil {
			log.Errorf(ctx, "failed GitHubNotifyStore.Put %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-type", "text-plain")
	w.Write(body)
}

func buildMessage(n *GitHubNotifyEntity) string {
	u := strings.Replace(n.URL, "api.github.com/repos", "github.com", -1)
	u = strings.Replace(u, "pulls", "pull", -1)
	return fmt.Sprintf("%s %s", n.Title, u)
}
