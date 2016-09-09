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
	"strconv"
	"strings"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/noaa/consumer"
)

const GROUP_CLOUD_CONTROLLER_ADMIN = "cloud_controller.admin"

type client struct {
	ClientId             string   `json:"client_id"`
	ClientSecret         string   `json:"client_secret,omitempty"`
	Scope                []string `json:"scope"`
	ResourceIds          []string `json:"resource_ids"`
	Authorities          []string `json:"authorities"`
	AuthorizedGrantTypes []string `json:"authorized_grant_types"`
}

type resourceSet struct {
	Resources    []map[string]interface{} `json:"resources"`
	TotalResults int                      `json:"totalResults"`
}

type groupResourceSet struct {
	Resources    []*group `json:"resources"`
	TotalResults int      `json:"totalResults"`
}

type user struct {
	UserName string  `json:"userName"`
	Origin   string  `json:"origin"`
	Emails   []email `json:"emails"`
}

type email struct {
	Value string `json:"value"`
}

type group struct {
	Meta        meta          `json:"meta"`
	DisplayName string        `json:"displayName"`
	Schema      []string      `json:"schemas"`
	Members     []groupMember `json:"members"`
	ZoneId      string        `json:"zoneId"`
	Id          string        `json:"id"`
}

type groupMember struct {
	Origin string `json:"origin"`
	Type   string `json:"type"`
	Value  string `json:"value"`
}

type meta struct {
	Version      int    `json:"version"`
	Created      string `json:"created"`
	LastModified string `json:"lastModified"`
}

type UaaRegistrar interface {
	RegisterFirehoseClient(uaaFirehoseUser string, uaaFirehoseSecret string) error
	RegisterUser(uaaFirehoseUser string, uaaFirehosePassword string) (string, error)
	AddUserToGroup(userId string, groupName string) error
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

func (p *uaaRegistrar) RegisterUser(uaaUser string, uaaPassword string) (string, error) {
	id, err := p.getUserId(uaaUser)
	if err != nil {
		return "", err
	}

	if id != "" {
		p.logger.Info(fmt.Sprintf("User [%s] already exists", uaaUser))
	} else {
		p.logger.Info(fmt.Sprintf("User [%s] doesn't exists, creating", uaaUser))
		id, err = p.createUser(uaaUser, uaaPassword)
	}
	if err != nil {
		return "", err
	}
	p.logger.Info(fmt.Sprintf("User id: %s", id))

	p.logger.Info("Setting firehose password")
	err = p.setPassword(id, uaaPassword)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (p *uaaRegistrar) AddUserToGroup(userId string, groupName string) error {
	group, err := p.getGroup(groupName)
	if err != nil {
		return err
	}

	for _, member := range group.Members {
		if member.Value == userId {
			p.logger.Info(fmt.Sprintf("User [%s] is already a member of [%s]", userId, groupName))
			return nil
		}
	}

	group.Members = append(group.Members, groupMember{
		Value:  userId,
		Origin: "uaa",
		Type:   "USER",
	})

	err = p.saveGroup(group)
	if err != nil {
		return err
	} else {
		p.logger.Info(fmt.Sprintf("Added [%s] to %s", userId, groupName))
		return nil
	}
}

func (p *uaaRegistrar) getUserId(uaaFirehoseUser string) (string, error) {
	url := fmt.Sprintf(`%s/Users?filter=userName+eq+"%s"`, p.uaaUrl, uaaFirehoseUser)
	resp, err := p.makeUaaRequest("GET", url, nil, map[string]string{})
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
			id := resourceSet.Resources[0]["id"].(string)
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
	resp, err := p.makeUaaRequest("POST", url, user, map[string]string{})
	if err != nil {
		return "", err
	}

	code := resp.StatusCode
	if code != 201 {
		responseBody, _ := ioutil.ReadAll(resp.Body)
		return "", errors.New(fmt.Sprintf("Creating user responded with %d:\n%+v", resp.StatusCode, string(responseBody)))
	} else {
		created := map[string]interface{}{}
		rawResponse, err := p.readAndUnmarshall(resp, &created)
		if err != nil {
			return "", err
		}

		id := created["id"]
		if id == nil {
			return "", errors.New(fmt.Sprintf("Couldn't parse create response:\n%+v", string(rawResponse)))
		} else {
			return id.(string), nil
		}
	}
}

func (p *uaaRegistrar) setPassword(uaaFirehoseUserId string, uaaFirehosePassword string) error {
	url := fmt.Sprintf("%s/Users/%s/password", p.uaaUrl, uaaFirehoseUserId)
	password := map[string]string{
		"password": uaaFirehosePassword,
	}
	resp, err := p.makeUaaRequest("PUT", url, password, map[string]string{})
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

func (p *uaaRegistrar) getGroup(displayName string) (*group, error) {
	url := fmt.Sprintf(`%s/Groups?filter=displayName+eq+"%s"`, p.uaaUrl, displayName)
	resp, err := p.makeUaaRequest("GET", url, nil, map[string]string{})
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(
			fmt.Sprintf("Get group responded with [%d]: %+v", resp.StatusCode, string(body)),
		)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		groups := groupResourceSet{}
		err := json.Unmarshal(body, &groups)
		if err != nil {
			return nil, err
		}
		if groups.TotalResults != 1 {
			return nil, errors.New(
				fmt.Sprintf("Expected a single group to match [%s]: %+v", displayName, string(body)),
			)
		}

		group := groups.Resources[0]
		//No way to fail on unexpected json keys ∴ re-serialize protect against schema changes
		check, err := json.Marshal(group)
		if err != nil {
			return nil, err
		}
		if !strings.Contains(string(body), string(check)) {
			return nil, errors.New(fmt.Sprintf(
				"UAA response schema didn't match expectations, response vs re-serialized:\n%s\n%s\n",
				string(body), string(check),
			))
		}
		return group, nil
	}
}

func (p *uaaRegistrar) saveGroup(group *group) error {
	url := fmt.Sprintf(`%s/Groups/%s`, p.uaaUrl, group.Id)

	resp, err := p.makeUaaRequest("PUT", url, group, map[string]string{
		"If-Match": strconv.Itoa(group.Meta.Version),
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(
			fmt.Sprintf("Save group responded with [%d]: %+v", resp.StatusCode, string(body)),
		)
	} else {
		return nil
	}
}

func (p *uaaRegistrar) clientExists(uaaFirehoseClient string) (bool, error) {
	url := fmt.Sprintf("%s/oauth/clients/%s", p.uaaUrl, uaaFirehoseClient)
	resp, err := p.makeUaaRequest("GET", url, nil, map[string]string{})
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
	resp, err := p.makeUaaRequest("POST", url, client, map[string]string{})
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
	resp, err := p.makeUaaRequest("PUT", reqUserUrl, client, map[string]string{})
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
	resp, err = p.makeUaaRequest("PUT", reqSecretUrl, secret, map[string]string{})
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

func (p *uaaRegistrar) makeUaaRequest(method string, url string, body interface{}, headers map[string]string) (*http.Response, error) {
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
	for key, value := range headers {
		req.Header.Add(key, value)
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
