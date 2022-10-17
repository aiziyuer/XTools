package global

import "sync"

var (
	onceRuntimeConfig     sync.Once
	instanceRuntimeConfig *RuntimeConfig
)

type RuntimeConfig struct {
	appName    string
	runtimeMap map[string]interface{}
}

func (t *RuntimeConfig) SetAppName(appName string) {
	t.appName = appName
}

func (t *RuntimeConfig) GetAppName() string {
	return t.appName
}

func AppConfig() *RuntimeConfig {

	onceRuntimeConfig.Do(func() {

		instanceRuntimeConfig = &RuntimeConfig{
			appName:    "",
			runtimeMap: make(map[string]interface{}, 0),
		}

	})

	return instanceRuntimeConfig
}

func (t *RuntimeConfig) Get(name string) interface{} {
	return t.runtimeMap[name]
}

func (t *RuntimeConfig) Set(name string, value interface{}) {
	t.runtimeMap[name] = value
}
