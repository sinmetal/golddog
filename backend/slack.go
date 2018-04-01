package backend

import (
	"context"

	"github.com/lestrrat-go/slack"
	"google.golang.org/appengine/urlfetch"
)

// PostMessage is SlackにMessageをPostする
func PostMessage(ctx context.Context, text string) error {
	ac := GetAppConfig(ctx)

	sl := slack.New(ac.SlackAPIKey, slack.WithClient(urlfetch.Client(ctx)))
	_, err := sl.Chat().PostMessage("sinmetal-github-nty").Text(text).Do(ctx)
	return err
}
