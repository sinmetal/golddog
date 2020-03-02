package backend

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.mercari.io/datastore"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func CronNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	ac := GetAppConfig(ctx)

	gc := NewGitHubClient(ctx, ac.GitHubToken)
	ns, err := gc.ListNotifications(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "text-plain")
		fmt.Printf("failed GitHub.ListNotifications. err=%v\n", err)
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
		key := store.Key(n.GetID())
		e, err := store.Get(ctx, key)
		if err == datastore.ErrNoSuchEntity {
			e = &GitHubNotifyEntity{
				ID:          n.GetID(),
				NotifyCount: 0,
				CreatedAt:   time.Now(),
			}
		} else if err != nil {
			log.Errorf(ctx, "%+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("%+v\n", e)
		t := e.NotifyAt.Add(time.Duration(e.NotifyCount) * time.Minute * 60)
		if e.NotifyCount > 0 && t.After(time.Now()) {
			log.Infof(ctx, "not snooze...")
			continue
		}

		e.Reason = n.GetReason()
		e.Title = n.GetSubject().GetTitle()
		e.URL = n.GetSubject().GetURL()
		e.LatestCommentURL = n.GetSubject().GetLatestCommentURL()
		e.Type = n.GetSubject().GetType()
		e.UpdatedAt = n.GetUpdatedAt()

		msg, err := buildMessage(e)
		if err != nil {
			log.Errorf(ctx, "failed buildMessage %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := PostMessage(ctx, msg); err != nil {
			log.Errorf(ctx, "failed slack.post %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		e.NotifyCount++
		e.NotifyAt = time.Now()
		_, err = store.Put(ctx, e)
		if err != nil {
			log.Errorf(ctx, "failed GitHubNotifyStore.Put %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func buildMessage(n *GitHubNotifyEntity) (string, error) {
	u := strings.Replace(n.URL, "api.github.com/repos", "github.com", -1)
	u = strings.Replace(u, "pulls", "pull", -1)

	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return "", err
	}

	jst := n.UpdatedAt.In(tokyo)
	return fmt.Sprintf("[%s:%s:%s][%s] %s %s", n.ID, n.Type, n.Reason, jst.Format("01-02 15:04"), n.Title, u), nil
}
