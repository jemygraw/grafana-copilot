package conf

import (
	"encoding/json"
	"fmt"
	"os"
)

var AppConfig *Config

type Config struct {
	// Grafana Host should be the URL of your Grafana instance, e.g. http://localhost:3000
	GrafanaHost string `json:"GRAFANA_HOST"`
	// Create the grafana token by adding a service account at http://localhost:3000/org/serviceaccounts
	GrafanaToken                string `json:"GRAFANA_TOKEN"`
	InfoflowRobotWebhookAddress string `json:"INFOFLOW_ROBOT_WEBHOOK_ADDRESS"`
	InfoflowRobotToken          string `json:"INFOFLOW_ROBOT_TOKEN"`
	InfoflowRobotEncodingAESKey string `json:"INFOFLOW_ROBOT_ENCODING_AES_KEY"`
	OpenAIAPIKey                string `json:"OPENAI_API_KEY"`
	OpenAIAPIBase               string `json:"OPENAI_API_BASE"`
	OpenAIModel                 string `json:"OPENAI_MODEL"`
}

func MustParseConfigFromEnvs() {
	appConfigMap := make(map[string]string)
	ensureEnv(&appConfigMap, "GRAFANA_HOST")
	ensureEnv(&appConfigMap, "GRAFANA_TOKEN")
	ensureEnv(&appConfigMap, "INFOFLOW_ROBOT_WEBHOOK_ADDRESS")
	ensureEnv(&appConfigMap, "INFOFLOW_ROBOT_TOKEN")
	ensureEnv(&appConfigMap, "INFOFLOW_ROBOT_ENCODING_AES_KEY")
	ensureEnv(&appConfigMap, "OPENAI_API_KEY")
	ensureEnv(&appConfigMap, "OPENAI_API_BASE")
	ensureEnv(&appConfigMap, "OPENAI_MODEL")
	appConfigData, _ := json.Marshal(appConfigMap)
	var res Config
	_ = json.Unmarshal(appConfigData, &res)
	AppConfig = &res
}

func ensureEnv(appConfigMap *map[string]string, key string) {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("Environment variable `%s` not set", key))
	}
	(*appConfigMap)[key] = value
}
