package cache

type NoCache struct{}

func NewNoCache() Cache {
	return &NoCache{}
}

func (c *NoCache) Open() error {
	return nil
}

func (c *NoCache) Close() error {
	return nil
}

func (c *NoCache) GetAllApps() (map[string]*App, error) {
	return nil, nil
}

func (c *NoCache) GetApp(appGuid string) (*App, error) {
	return nil, nil
}
