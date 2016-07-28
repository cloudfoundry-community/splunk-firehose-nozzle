package auth

type AuthTokenFetcher interface {
	FetchAuthToken() (string, error)
}
