package testing

type TokenClientMock struct {
	GetTokenFn func() (string, error)
}

func (m *TokenClientMock) GetToken() (string, error) {
	if m.GetTokenFn != nil {
		return m.GetTokenFn()
	} else {
		return "", nil
	}
}
