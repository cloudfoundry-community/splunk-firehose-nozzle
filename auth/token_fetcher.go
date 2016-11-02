package auth

import (
	"errors"

	"github.com/cloudfoundry/noaa/consumer"
)

type TokenGetter interface {
	GetToken() (string, error)
}

type TokenRefreshAdapter struct {
	tokenGetter TokenGetter
}

func NewTokenRefreshAdapter(tokenGetter TokenGetter) consumer.TokenRefresher {
	return &TokenRefreshAdapter{
		tokenGetter: tokenGetter,
	}
}

func (t *TokenRefreshAdapter) RefreshAuthToken() (string, error) {
	token, err := t.tokenGetter.GetToken()
	if err != nil {
		return "", err
	} else if token == "" {
		return "", errors.New("TokenGetter failed to return a token")
	} else {
		return token, nil
	}
}
