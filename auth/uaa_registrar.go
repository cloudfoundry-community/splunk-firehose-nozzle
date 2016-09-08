package auth

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/noaa/consumer"
)

type client struct {
	ClientId             string   `json:"client_id"`
	ClientSecret         string   `json:"client_secret,omitempty"`
	Scope                []string `json:"scope"`
	ResourceIds          []string `json:"resource_ids"`
	Authorities          []string `json:"authorities"`
	AuthorizedGrantTypes []string `json:"authorized_grant_types"`
}

type resourceSet struct {
	Resources    []resource `json:"resources"`
	TotalResults int        `json:"totalResults"`
}

type user struct {
	UserName string  `json:"userName"`
	Origin   string  `json:"origin"`
	Emails   []email `json:"emails"`
}

type email struct {
	Value string `json:"value"`
}

type resource struct {
	Id string `json:"id"`
}

type UaaRegistrar interface {
	RegisterFirehoseClient(uaaFirehoseUser string, uaaFirehoseSecret string) error
	RegisterAdminUser(uaaFirehoseUser string, uaaFirehosePassword string) error
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

func (p *uaaRegistrar) RegisterFirehoseClient(uaaFirehoseClient string, uaaFirehoseSecret string) error {
	exists, err := p.clientExists(uaaFirehoseClient)
	if err != nil {
		return err
	}

	if exists {
		p.logger.Info("Firehose *client* exists, updating")
		return p.updateClient(uaaFirehoseClient, uaaFirehoseSecret)
	} else {
		p.logger.Info("Firehose *client* doesn't exists, creating")
		return p.createClient(uaaFirehoseClient, uaaFirehoseSecret)
	}
}

func (p *uaaRegistrar) RegisterAdminUser(uaaFirehoseUser string, uaaFirehoseSecret string) error {
	id, err := p.getUserId(uaaFirehoseUser)
	if err != nil {
		return err
	}

	if id != "" {
		p.logger.Info("Firehose *user* already exists")
	} else {
		p.logger.Info("Firehose *user* doesn't exists, creating")
		id, err = p.createUser(uaaFirehoseUser, uaaFirehoseSecret)
	}
	if err != nil {
		return err
	}
	p.logger.Info(fmt.Sprintf("Firehose user id: %s", id))

	p.logger.Info("Setting firehose password")
	err = p.setPassword(id, uaaFirehoseSecret)
	if err != nil {
		return err
	}

	return nil
}

func (p *uaaRegistrar) clientExists(uaaFirehoseClient string) (bool, error) {
	url := fmt.Sprintf("%s/oauth/clients/%s", p.uaaUrl, uaaFirehoseClient)
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

func (p *uaaRegistrar) getUserId(uaaFirehoseUser string) (string, error) {
	url := fmt.Sprintf(`%s/Users?filter=userName+eq+"%s"`, p.uaaUrl, uaaFirehoseUser)
	reqUser, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	reqUser.Header.Add("Authorization", p.authToken)
	resp, err := p.httpClient.Do(reqUser)
	if err != nil {
		return "", err
	}

	code := resp.StatusCode
	if code != 200 {
		return "", errors.New(fmt.Sprintf("Checking if user exists responded incorrectly: %+v", resp))
	} else {
		resourceSet := resourceSet{}
		_, err := p.readAndUnmarshall(resp, &resourceSet)
		if err != nil {
			return "", err
		}

		if resourceSet.TotalResults == 0 {
			return "", nil
		} else if resourceSet.TotalResults == 1 {
			id := resourceSet.Resources[0].Id
			return id, nil
		} else {
			return "", errors.New(fmt.Sprintf("Checking if user exists responded with more than 1 user:\n%+v", resourceSet))
		}
	}
}

func (p *uaaRegistrar) createUser(uaaFirehoseUser string, uaaFirehosePassword string) (string, error) {
	url := fmt.Sprintf(`%s/Users`, p.uaaUrl)
	requestBody, err := json.Marshal(user{
		UserName: uaaFirehoseUser,
		Origin:   "uaa",
		Emails: []email{
			{Value: uaaFirehoseUser},
		},
	})
	if err != nil {
		return "", err
	}

	reqUser, err := http.NewRequest("POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return "", err
	}
	reqUser.Header.Add("Authorization", p.authToken)
	reqUser.Header.Add("Content-Type", "application/json")
	resp, err := p.httpClient.Do(reqUser)
	if err != nil {
		return "", err
	}

	code := resp.StatusCode
	if code != 201 {
		responseBody, _ := ioutil.ReadAll(resp.Body)
		return "", errors.New(fmt.Sprintf("Creating user responded with %d:\n%+v", resp.StatusCode, string(responseBody)))
	} else {
		created := resource{}
		rawResponse, err := p.readAndUnmarshall(resp, &created)
		if err != nil {
			return "", err
		}

		id := created.Id
		if id == "" {
			return "", errors.New(fmt.Sprintf("Couldn't parse create response:\n%+v", string(rawResponse)))
		} else {
			return id, nil
		}
	}
}

func (p *uaaRegistrar) setPassword(uaaFirehoseUserId string, uaaFirehosePassword string) error {
	url := fmt.Sprintf("%s/Users/%s/password", p.uaaUrl, uaaFirehoseUserId)
	body, err := json.Marshal(map[string]string{
		"password": uaaFirehosePassword,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", p.authToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(
			fmt.Sprintf("Update user password responded with [%d]: %+v", resp.StatusCode, body),
		)
	} else {
		return nil
	}
}

func (p *uaaRegistrar) setAdmin(uaaFirehoseUser string) error {
	panic("Todo")
	return nil
}

func (p *uaaRegistrar) createClient(uaaFirehoseUser string, uaaFirehoseSecret string) error {
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
		return err
	}
	reqUser.Header.Add("Authorization", p.authToken)
	reqUser.Header.Add("Content-Type", "application/json")

	resp, err := p.httpClient.Do(reqUser)

	if resp.StatusCode != 201 {
		return errors.New(fmt.Sprintf("Create client responded incorrectly: %+v", resp))
	} else {
		return nil
	}
}

func (p *uaaRegistrar) updateClient(uaaFirehoseUser string, uaaFirehoseSecret string) error {
	reqUserUrl := fmt.Sprintf("%s/oauth/clients/%s", p.uaaUrl, uaaFirehoseUser)
	client := p.getFirehoseClient()
	client.ClientId = uaaFirehoseUser
	body, err := json.Marshal(client)
	if err != nil {
		return err
	}

	reqUser, err := http.NewRequest("PUT", reqUserUrl, bytes.NewReader(body))
	if err != nil {
		return err
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
	} else {
		return nil
	}
}

func (p *uaaRegistrar) readAndUnmarshall(resp *http.Response, target interface{}) (string, error) {
	rawResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(rawResponse, target)
	if err != nil {
		return string(rawResponse), err
	} else {
		return string(rawResponse), nil
	}
}

func (p *uaaRegistrar) getFirehoseClient() *client {
	return &client{
		Scope:                []string{"openid", "oauth.approvals", "doppler.firehose"},
		ResourceIds:          []string{"none"},
		Authorities:          []string{"oauth.login", "doppler.firehose"},
		AuthorizedGrantTypes: []string{"client_credentials"},
	}
}
