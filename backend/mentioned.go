package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", ac.GitHubToken))
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Infof(ctx, "%s", string(body))
	var ns []GitHubNotification
	if err := json.Unmarshal(body, &ns); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(ns) < 1 {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text-plain")
		w.Write([]byte("nothing update"))
		return
	}
	msg := buildMessage(&ns[0])

	_, err = putGitHubNotify(ctx, &ns[0])
	if err != nil {
		log.Errorf(ctx, "failed datastore.putGitHubNotify %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := PostMessage(ctx, msg); err != nil {
		log.Errorf(ctx, "failed slack.post %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "text-plain")
	w.Write(body)
}

func buildMessage(n *GitHubNotification) string {
	u := strings.Replace(n.Subject.URL, "api.github.com/repos", "github.com", -1)
	u = strings.Replace(u, "pulls", "pull", -1)
	return fmt.Sprintf("%s %s", n.Subject.Title, u)
}

// GitHubNotifyEntity is GitHubNotifyをDatastoreに保存するためのEntity
type GitHubNotifyEntity struct {
	ID               string `json:"id" datastore:"-"`
	Title            string `json:"title"`
	URL              string `json:"url"`
	LatestCommentURL string `json:"latest_comment_url"`
	Type             string `json:"type"`
}

func putGitHubNotify(ctx context.Context, n *GitHubNotification) (*GitHubNotifyEntity, error) {
	ds, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}
	k := ds.NameKey("GitHubNotify", n.ID, nil)
	e := GitHubNotifyEntity{
		ID:               n.ID,
		Title:            n.Subject.Title,
		URL:              n.Subject.URL,
		LatestCommentURL: n.Subject.LatestCommentURL,
		Type:             n.Subject.Type,
	}
	_, err = ds.Put(ctx, k, &e)
	if err != nil {
		return nil, err
	}
	return &e, nil
}
