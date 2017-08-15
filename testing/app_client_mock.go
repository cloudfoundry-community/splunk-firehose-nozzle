package testing

import (
	"errors"
	"fmt"
	"net/url"
	"sync"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

type AppClientMock struct {
	lock sync.RWMutex
	apps map[string]cfclient.App
	n    int
}

func NewAppClientMock(n int) *AppClientMock {
	apps := getApps(n)
	return &AppClientMock{
		apps: apps,
		n:    n,
	}
}

func (m *AppClientMock) AppByGuid(guid string) (cfclient.App, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	app, ok := m.apps[guid]
	if ok {
		return app, nil
	}
	return app, errors.New("No such app")
}

func (m *AppClientMock) ListApps() ([]cfclient.App, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var apps []cfclient.App
	for k := range m.apps {
		apps = append(apps, m.apps[k])
	}
	return apps, nil
}

func (m *AppClientMock) ListAppsByQueryWithLimits(query url.Values, totalPages int) ([]cfclient.App, error) {
	if totalPages <= 0 {
		return m.ListApps()
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	var apps []cfclient.App
	count := 0
	for k := range m.apps {
		if count >= totalPages*100 {
			break
		}
		count += 1
		apps = append(apps, m.apps[k])
	}
	return apps, nil
}

func (m *AppClientMock) CreateApp(appID, spaceID, orgID string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	app := cfclient.App{
		Guid: appID,
		Name: appID,
		SpaceData: cfclient.SpaceResource{
			Entity: cfclient.Space{
				Guid: spaceID,
				Name: spaceID,
				OrgData: cfclient.OrgResource{
					Entity: cfclient.Org{
						Guid: orgID,
						Name: orgID,
					},
				},
			},
		},
	}

	m.apps[appID] = app
}

func getApps(n int) map[string]cfclient.App {
	apps := make(map[string]cfclient.App, n)
	for i := 0; i < n; i++ {
		app := cfclient.App{
			Guid: fmt.Sprintf("cf_app_id_%d", i),
			Name: fmt.Sprintf("cf_app_name_%d", i),
			SpaceData: cfclient.SpaceResource{
				Entity: cfclient.Space{
					Guid: fmt.Sprintf("cf_space_id_%d", i%50),
					Name: fmt.Sprintf("cf_space_name_%d", i%50),
					OrgData: cfclient.OrgResource{
						Entity: cfclient.Org{
							Guid: fmt.Sprintf("cf_org_id_%d", i%100),
							Name: fmt.Sprintf("cf_org_name_%d", i%100),
						},
					},
				},
			},
		}
		apps[app.Guid] = app
	}
	return apps
}
