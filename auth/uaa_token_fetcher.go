package auth

import "github.com/cloudfoundry-incubator/uaago"

type uaaTokenFetcher struct {
	uaaUrl                string
	username              string
	password              string
	insecureSSLSkipVerify bool
}

func NewUAATokenFetcher(url string, username string, password string, sslSkipVerify bool) *uaaTokenFetcher {
	return &uaaTokenFetcher{
		uaaUrl:                url,
		username:              username,
		password:              password,
		insecureSSLSkipVerify: sslSkipVerify,
	}
}

func (uaa *uaaTokenFetcher) FetchAuthToken() (string, error) {
	uaaClient, err := uaago.NewClient(uaa.uaaUrl)
	if err != nil {
		return "", err
	}

	authToken, err := uaaClient.GetAuthToken(uaa.username, uaa.password, uaa.insecureSSLSkipVerify)
	if err != nil {
		return "", err
	}
	return authToken, nil
}
