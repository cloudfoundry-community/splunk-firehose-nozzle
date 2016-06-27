package nozzle

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cloudfoundry-incubator/cf_http"
)

type SplunkEvent struct {
	Time       string `json:"time,omitempty"`
	Host       string `json:"host,omitempty"`
	Source     string `json:"source,omitempty"`
	SourceType string `json:"sourcetype,omitempty"`
	Index      string `json:"index,omitempty"`

	Event interface{} `json:"event"`
}

type SplunkClient interface {
	Post(*SplunkEvent) error
}

type splunkClient struct {
	httpClient  *http.Client
	splunkToken string
	splunkHost  string
}

func NewSplunkClient(splunkToken string, splunkHost string, insecureSkipVerify bool) SplunkClient {
	httpClient := cf_http.NewClient()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}
	httpClient.Transport = tr

	return &splunkClient{
		httpClient:  httpClient,
		splunkToken: splunkToken,
		splunkHost:  splunkHost,
	}
}

func (s *splunkClient) Post(event *SplunkEvent) error {
	postBody, err := json.Marshal(event)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/services/collector", s.splunkHost)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(postBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Splunk %s", s.splunkToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		responseBody, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Non-ok response code [%d] from splunk: %s", resp.StatusCode, responseBody))
	}

	return nil
}
