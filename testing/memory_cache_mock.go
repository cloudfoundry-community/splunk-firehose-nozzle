package testing

import (
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/cache"
)

type MemoryCacheMock struct {
	ignoreApp bool
}

func NewMemoryCacheMock() *MemoryCacheMock {
	return &MemoryCacheMock{}
}

func (c *MemoryCacheMock) Open() error {
	return nil
}

func (c *MemoryCacheMock) Close() error {
	return nil
}

func (c *MemoryCacheMock) GetAllApps() (map[string]*cache.App, error) {
	return nil, nil
}

func (c *MemoryCacheMock) GetApp(appGuid string) (*cache.App, error) {
	app := &cache.App{
		Name:       "testing-app",
		Guid:       "f964a41c-76ac-42c1-b2ba-663da3ec22d5",
		SpaceName:  "testing-space",
		SpaceGuid:  "f964a41c-76ac-42c1-b2ba-663da3ec22d6",
		OrgName:    "testing-org",
		OrgGuid:    "f964a41c-76ac-42c1-b2ba-663da3ec22d7",
		IgnoredApp: c.ignoreApp,
	}

	return app, nil
}

func (c *MemoryCacheMock) SetIgnoreApp(ignore bool) {
	c.ignoreApp = ignore
}
