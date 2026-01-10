package conf

import (
	"encoding/json"
	"fmt"
	"os"
)

var AppConfig *Config

type Config struct {
	// GrafanaHost should be the api host of your Grafana instance which is used to create openapi requests
	// e.g. http://localhost:3000
	GrafanaHost string `json:"GRAFANA_HOST"`
	// GrafanaBaseURL should be the base URL of your Grafana dashboard which is used to create full access URL
	// It is default to GrafanaHost if not set.
	// e.g. http://localhost:3000
	GrafanaBaseURL string `json:"GRAFANA_BASE_URL"`
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
	// check grafana base url
	if os.Getenv("GRAFANA_BASE_URL") != "" {
		appConfigMap["GRAFANA_BASE_URL"] = os.Getenv("GRAFANA_BASE_URL")
	} else {
		appConfigMap["GRAFANA_BASE_URL"] = appConfigMap["GRAFANA_HOST"]
	}
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
