package testing

type MockSplunkClient struct {
	CapturedEvents []map[string]interface{}
	PostBatchFn    func(events []map[string]interface{}) error
}

func (m *MockSplunkClient) Post(events []map[string]interface{}) error {
	if m.PostBatchFn != nil {
		return m.PostBatchFn(events)
	} else {
		m.CapturedEvents = append(m.CapturedEvents, events...)
	}
	return nil
}

type MockTokenGetter struct {
	GetTokenFn func() string
}

func (m *MockTokenGetter) GetToken() string {
	if m.GetTokenFn != nil {
		return m.GetTokenFn()
	} else {
		return ""
	}
}
