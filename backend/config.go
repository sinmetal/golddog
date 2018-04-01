package backend

import (
	"context"
	"sync"

	"google.golang.org/appengine/log"
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
		ac, err := getAppConfigFromDatastore(ctx)
		if err != nil {
			// FIXME errが帰ってこなかった場合、defaultConfigが初期化されず大きな悲しみに包まれる
			// https://github.com/gcpug/nouhau/issues/31 がいい感じになるのを待つ
			log.Errorf(ctx, "大いなる悲しみのエラー %+v", err)
			return
		}
		defaultConfig = ac
	})
	return defaultConfig
}

// getAppConfigImple
func getAppConfigFromDatastore(ctx context.Context) (*AppConfig, error) {
	ds, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}
	k := ds.NameKey("AppConfig", "golddog", nil)
	var c AppConfig
	if err := ds.Get(ctx, k, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
