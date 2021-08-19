package backend

import (
	"context"
	"os"
	"sync"
)

var defaultConfig *AppConfig

// AppConfig is Application Config
type AppConfig struct {
	SlackAPIKey string
	GitHubToken string
}

// GetAppConfig is Application Configを取得する
func GetAppConfig(ctx context.Context) *AppConfig {
	var once sync.Once

	once.Do(func() {
		defaultConfig = &AppConfig{
			SlackAPIKey: os.Getenv("SLACK_TOKEN"),
			GitHubToken: os.Getenv("GITHUB_TOKEN"),
		}
	})
	return defaultConfig
}
