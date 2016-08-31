package auth

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/noaa/consumer"
)

type Client struct {
	ClientId             string   `json:"client_id"`
	ClientSecret         string   `json:"client_secret,omitempty"`
	Scope                []string `json:"scope"`
	ResourceIds          []string `json:"resource_ids"`
	Authorities          []string `json:"authorities"`
	AuthorizedGrantTypes []string `json:"authorized_grant_types"`
}

type UaaRegistrar interface {
	RegisterFirehose(uaaFirehoseUser string, uaaFirehoseSecret string) error
}

type uaaRegistrar struct {
	httpClient *http.Client
	uaaUrl     string
	authToken  string
	logger     lager.Logger
}

func NewUaaRegistrar(uaaUrl string, tokenRefresher consumer.TokenRefresher, insecureSkipVerify bool, logger lager.Logger) (UaaRegistrar, error) {
	authToken, err := tokenRefresher.RefreshAuthToken()
	if err != nil {
		return nil, err
	}

	config := &tls.Config{InsecureSkipVerify: insecureSkipVerify}
	transport := &http.Transport{TLSClientConfig: config}
	httpClient := &http.Client{Transport: transport}

	return &uaaRegistrar{
		httpClient: httpClient,
		uaaUrl:     uaaUrl,
		authToken:  authToken,
		logger:     logger,
	}, nil
}

func (p *uaaRegistrar) RegisterFirehose(uaaFirehoseUser string, uaaFirehoseSecret string) error {
	exists, err := p.exists(uaaFirehoseUser)
	if err != nil {
		return err
	}

	if exists {
		return p.update(uaaFirehoseUser, uaaFirehoseSecret)
	} else {
		return p.create(uaaFirehoseUser, uaaFirehoseSecret)
	}
}

func (p *uaaRegistrar) exists(uaaFirehoseUser string) (bool, error) {
	url := fmt.Sprintf("%s/oauth/clients/%s", p.uaaUrl, uaaFirehoseUser)
	reqUser, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	reqUser.Header.Add("Authorization", p.authToken)
	resp, err := p.httpClient.Do(reqUser)
	if err != nil {
		return false, err
	}

	code := resp.StatusCode
	if code == 200 {
		return true, nil
	} else if code == 404 {
		return false, nil
	} else {
		return false, errors.New(fmt.Sprintf("Checking if client exists responded incorrectly: %+v", resp))
	}
}

func (p *uaaRegistrar) create(uaaFirehoseUser string, uaaFirehoseSecret string) error {
	url := fmt.Sprintf("%s/oauth/clients", p.uaaUrl)
	client := p.getFirehoseClient()
	client.ClientId = uaaFirehoseUser
	client.ClientSecret = uaaFirehoseSecret
	body, err := json.Marshal(client)
	if err != nil {
		return err
	}

	reqUser, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	reqUser.Header.Add("Authorization", p.authToken)
	reqUser.Header.Add("Content-Type", "application/json")

	resp, err := p.httpClient.Do(reqUser)

	if resp.StatusCode != 201 {
		return errors.New(fmt.Sprintf("Create client responded incorrectly: %+v", resp))
	}
	return nil
}

func (p *uaaRegistrar) update(uaaFirehoseUser string, uaaFirehoseSecret string) error {
	reqUserUrl := fmt.Sprintf("%s/oauth/clients/%s", p.uaaUrl, uaaFirehoseUser)
	client := p.getFirehoseClient()
	client.ClientId = uaaFirehoseUser
	body, err := json.Marshal(client)
	if err != nil {
		return err
	}

	reqUser, err := http.NewRequest("PUT", reqUserUrl, bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	reqUser.Header.Add("Authorization", p.authToken)
	reqUser.Header.Add("Content-Type", "application/json")

	resp, err := p.httpClient.Do(reqUser)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Update client responded incorrectly: %+v", resp))
	}

	p.logger.Info(fmt.Sprintf("Update resp: %d", resp.StatusCode))

	reqSecretUrl := fmt.Sprintf("%s/oauth/clients/%s/secret", p.uaaUrl, uaaFirehoseUser)
	body, err = json.Marshal(map[string]string{
		"secret": uaaFirehoseSecret,
	})
	if err != nil {
		return err
	}

	reqSecret, err := http.NewRequest("PUT", reqSecretUrl, bytes.NewReader(body))
	if err != nil {
		return err
	}
	reqSecret.Header.Add("Authorization", p.authToken)
	reqSecret.Header.Add("Content-Type", "application/json")

	resp, err = p.httpClient.Do(reqSecret)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Update client secret responded incorrectly: %+v", resp))
	}

	return nil
}

func (p *uaaRegistrar) getFirehoseClient() *Client {
	return &Client{
		Scope:                []string{"openid", "oauth.approvals", "doppler.firehose"},
		ResourceIds:          []string{"none"},
		Authorities:          []string{"oauth.login", "doppler.firehose"},
		AuthorizedGrantTypes: []string{"client_credentials"},
	}
}
