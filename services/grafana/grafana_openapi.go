package grafana

import (
	"encoding/json"
	"fmt"
	"github.com/jemygraw/grafana-copilot/conf"
	"net/http"
	"net/url"
	"time"
)

var grafanaClient = http.Client{
	Timeout: time.Second * 10,
}

type Dashboard struct {
	Uid   string `json:"uid"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// ListDashboardMeta list dashboards using grafana dashboard query api.
// See https://grafana.com/docs/grafana/latest/developer-resources/api-reference/http-api/folder_dashboard_search/
func ListDashboardMeta(query string) (dashboardList []Dashboard, err error) {
	reqParams := url.Values{}
	reqParams.Add("query", query)
	reqParams.Add("type", "dash-db")
	reqURL := fmt.Sprintf("%s/api/search?%s", conf.AppConfig.GrafanaHost, reqParams.Encode())
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		err = fmt.Errorf("new grafana request err: %v", err)
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", conf.AppConfig.GrafanaToken))
	resp, err := grafanaClient.Do(req)
	if err != nil {
		err = fmt.Errorf("call grafana api err, %s", err.Error())
		return
	}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&dashboardList); err != nil {
		err = fmt.Errorf("decode grafana api resp err, %s", err.Error())
		return
	}
	return
}
