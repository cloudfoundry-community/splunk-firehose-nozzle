package testing

import (
	"errors"
	"fmt"
	"net/url"
	"sync"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

type AppClientMock struct {
	lock                    sync.RWMutex
	apps                    map[string]cfclient.App
	n                       int
	listAppsCallCount       int
	appByGUIDCallCount      int
	getOrgByGUIDCallCount   int
	getSpaceByGUIDCallCount int
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

	m.appByGUIDCallCount++

	app, ok := m.apps[guid]
	if ok {
		return app, nil
	}
	return app, errors.New("No such app")
}

func (m *AppClientMock) ListApps() ([]cfclient.App, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	m.listAppsCallCount++

	var apps []cfclient.App
	for k := range m.apps {
		apps = append(apps, m.apps[k])
	}
	return apps, nil
}

func (m *AppClientMock) ListAppsByQueryWithLimits(query url.Values, totalPages int) ([]cfclient.App, error) {
	return m.ListApps()
}

func (m *AppClientMock) GetSpaceByGuid(spaceGUID string) (cfclient.Space, error) {
	m.getSpaceByGUIDCallCount++

	var id int
	fmt.Sscanf(spaceGUID, "cf_space_id_%d", &id)

	return cfclient.Space{
		Guid:             spaceGUID,
		Name:             fmt.Sprintf("cf_space_name_%d", id),
		OrganizationGuid: fmt.Sprintf("cf_org_id_%d", id),
	}, nil
}

func (m *AppClientMock) GetOrgByGuid(orgGUID string) (cfclient.Org, error) {
	m.getOrgByGUIDCallCount++

	var id int
	fmt.Sscanf(orgGUID, "cf_org_id_%d", &id)

	return cfclient.Org{
		Guid: orgGUID,
		Name: fmt.Sprintf("cf_org_name_%d", id),
	}, nil
}

func (m *AppClientMock) CreateApp(appID, spaceID string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	app := cfclient.App{
		Guid:      appID,
		Name:      appID,
		SpaceGuid: spaceID,
	}

	m.apps[appID] = app
}

func getApps(n int) map[string]cfclient.App {
	apps := make(map[string]cfclient.App, n)
	for i := 0; i < n; i++ {
		app := cfclient.App{
			Guid:      fmt.Sprintf("cf_app_id_%d", i),
			Name:      fmt.Sprintf("cf_app_name_%d", i),
			SpaceGuid: fmt.Sprintf("cf_space_id_%d", i%50),
		}
		apps[app.Guid] = app
	}
	return apps
}

func (m *AppClientMock) ListAppsCallCount() int       { return m.listAppsCallCount }
func (m *AppClientMock) AppByGUIDCallCount() int      { return m.appByGUIDCallCount }
func (m *AppClientMock) GetOrgByGUIDCallCount() int   { return m.getOrgByGUIDCallCount }
func (m *AppClientMock) GetSpaceByGUIDCallCount() int { return m.getSpaceByGUIDCallCount }

func (m *AppClientMock) ResetCallCounts() {
	m.listAppsCallCount = 0
	m.appByGUIDCallCount = 0
	m.getOrgByGUIDCallCount = 0
	m.getSpaceByGUIDCallCount = 0
}
