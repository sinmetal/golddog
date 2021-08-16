package backend

import (
	"context"

	"github.com/lestrrat-go/slack"
)

// PostMessage is SlackにMessageをPostする
func PostMessage(ctx context.Context, text string) error {
	ac := GetAppConfig(ctx)

	sl := slack.New(ac.SlackAPIKey)
	_, err := sl.Chat().PostMessage("sinmetal-github-nty").Text(text).Do(ctx)
	return err
}
