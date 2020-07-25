package eventsource

import (
	"crypto/tls"
	"net/http"
)

// tokenFetcher implements GetAuthToken and fetches token.
type tokenFetcher interface {
	GetAuthToken(clientID, secret string, skipCertVerify bool) (string, error)
}

//AuthClient provides client config and token fetcher for authentication.
type AuthClient struct {
	tokenFetcher   tokenFetcher
	clientID       string
	secret         string
	skipCertVerify bool
	httpClient     *http.Client
}

// NewHttp returns http client.
func NewHttp(tf tokenFetcher, clientID, secret string, skipCertVerify bool) *AuthClient {
	return &AuthClient{
		tokenFetcher:   tf,
		clientID:       clientID,
		secret:         secret,
		skipCertVerify: skipCertVerify,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: skipCertVerify,
				},
			},
		},
	}
}

// Do sends http request and returns the response.
func (c *AuthClient) Do(req *http.Request) (*http.Response, error) {
	token, err := c.tokenFetcher.GetAuthToken(
		c.clientID,
		c.secret,
		c.skipCertVerify,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", token)
	return c.httpClient.Do(req)
}
