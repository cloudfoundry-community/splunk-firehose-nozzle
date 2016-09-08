package auth

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

	p.logger.Info("Adding to admin group")
	/*err = */ p.setAdmin(id)
	/*
		if err != nil {
			return err
		}
	*/

	return nil
}

func (p *uaaRegistrar) getUserId(uaaFirehoseUser string) (string, error) {
	url := fmt.Sprintf(`%s/Users?filter=userName+eq+"%s"`, p.uaaUrl, uaaFirehoseUser)
	resp, err := p.makeUaaRequest("GET", url, nil)
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
	user := user{
		UserName: uaaFirehoseUser,
		Origin:   "uaa",
		Emails: []email{
			{Value: uaaFirehoseUser},
		},
	}
	resp, err := p.makeUaaRequest("POST", url, user)
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
	password := map[string]string{
		"password": uaaFirehosePassword,
	}
	resp, err := p.makeUaaRequest("PUT", url, password)
	if err != nil {
		return err
	}

	//Undocumented response code 422 when submit same password
	if resp.StatusCode != 200 && resp.StatusCode != 422 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(
			fmt.Sprintf("Update user password responded with [%d]: %+v", resp.StatusCode, string(body)),
		)
	} else {
		return nil
	}
}

func (p *uaaRegistrar) setAdmin(uaaFirehoseUserId string) error {
	//filter groups by name to get id and members
	//put group with new member added

	return nil
}

func (p *uaaRegistrar) clientExists(uaaFirehoseClient string) (bool, error) {
	url := fmt.Sprintf("%s/oauth/clients/%s", p.uaaUrl, uaaFirehoseClient)
	resp, err := p.makeUaaRequest("GET", url, nil)
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

func (p *uaaRegistrar) createClient(uaaFirehoseUser string, uaaFirehoseSecret string) error {
	url := fmt.Sprintf("%s/oauth/clients", p.uaaUrl)
	client := p.getFirehoseClient()
	client.ClientId = uaaFirehoseUser
	client.ClientSecret = uaaFirehoseSecret
	resp, err := p.makeUaaRequest("POST", url, client)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Create client responded incorrectly [%d]: %s", resp.StatusCode, body))
	} else {
		return nil
	}
}

func (p *uaaRegistrar) updateClient(uaaFirehoseUser string, uaaFirehoseSecret string) error {
	reqUserUrl := fmt.Sprintf("%s/oauth/clients/%s", p.uaaUrl, uaaFirehoseUser)
	client := p.getFirehoseClient()
	client.ClientId = uaaFirehoseUser
	resp, err := p.makeUaaRequest("PUT", reqUserUrl, client)
	if err != nil {
		return err
	}
	p.logger.Info(fmt.Sprintf("Update resp: %d", resp.StatusCode))
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Update client responded incorrectly [%d]: %s", resp.StatusCode, body))
	}

	reqSecretUrl := fmt.Sprintf("%s/oauth/clients/%s/secret", p.uaaUrl, uaaFirehoseUser)
	secret := map[string]string{
		"secret": uaaFirehoseSecret,
	}
	resp, err = p.makeUaaRequest("PUT", reqSecretUrl, secret)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Update client secret responded incorrectly [%d]: %s", resp.StatusCode, body))
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

func (p *uaaRegistrar) makeUaaRequest(method string, url string, body interface{}) (*http.Response, error) {
	var requestBody io.Reader
	if body != nil {
		requestBodyJson, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewReader(requestBodyJson)
	}
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", p.authToken)
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (p *uaaRegistrar) getFirehoseClient() *client {
	return &client{
		Scope:                []string{"openid", "oauth.approvals", "doppler.firehose"},
		ResourceIds:          []string{"none"},
		Authorities:          []string{"oauth.login", "doppler.firehose"},
		AuthorizedGrantTypes: []string{"client_credentials"},
	}
}
