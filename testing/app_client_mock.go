package testing

import (
	"errors"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"net/url"
	"sync"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

type AppClientMock struct {
	lock                    sync.RWMutex
	apps                    map[string]*resource.App //map[string]cfclient.App //v3 v2
	n                       int
	listAppsCallCount       int
	appByGUIDCallCount      int
	getOrgByGUIDCallCount   int
	getSpaceByGUIDCallCount int
	CfClientVersion         string
}

func NewAppClientMock(n int, cfclientversion string) *AppClientMock {
	apps := getApps(n, cfclientversion)
	return &AppClientMock{
		apps:            apps,
		n:               n,
		CfClientVersion: cfclientversion,
	}
}

// v2
// func (m *AppClientMock) AppByGuid(guid string) (cfclient.App, error) {
// v3
func (m *AppClientMock) AppByGuid(guid string) (*resource.App, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	m.appByGUIDCallCount++

	app, ok := m.apps[guid]
	if ok {
		return app, nil
	}
	return app, errors.New("No such app")
}

// v2
// func (m *AppClientMock) ListApps() ([]cfclient.App, error) {
// v3
func (m *AppClientMock) ListApps() ([]*resource.App, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	m.listAppsCallCount++

	var apps []*resource.App
	for k := range m.apps {
		apps = append(apps, m.apps[k])
	}
	return apps, nil
}

// v2
//
//	func (m *AppClientMock) ListAppsByQueryWithLimits(query url.Values, totalPages int) ([]cfclient.App, error) {
//		return m.ListApps()
//	}
//
// v3
func (m *AppClientMock) ListAppsByQueryWithLimits(query url.Values, totalPages int) ([]*resource.App, error) {
	return m.ListApps()
}

// v2
// func (m *AppClientMock) GetSpaceByGuid(spaceGUID string) (cfclient.Space, error) {
// v3
func (m *AppClientMock) GetSpaceByGuid(spaceGUID string) (*resource.Space, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.getSpaceByGUIDCallCount++

	var id int
	fmt.Sscanf(spaceGUID, "cf_space_id_%d", &id)
	//v2
	if m.CfClientVersion == "V2" {
		return cfclient.Space{
			Guid:             spaceGUID,
			Name:             fmt.Sprintf("cf_space_name_%d", id),
			OrganizationGuid: fmt.Sprintf("cf_org_id_%d", id),
		}, nil
	}
	//v3
	return &resource.Space{
		Resource:      resource.Resource{GUID: spaceGUID},
		Name:          fmt.Sprintf("cf_space_name_%d", id),
		Relationships: &resource.SpaceRelationships{Organization: &resource.ToOneRelationship{Data: &resource.Relationship{GUID: fmt.Sprintf("cf_org_id_%d", id)}}},
	}, nil
}

// v2
// func (m *AppClientMock) GetOrgByGuid(orgGUID string) (cfclient.Org, error) {
// v3
func (m *AppClientMock) GetOrgByGuid(orgGUID string) (*resource.Organization, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.getOrgByGUIDCallCount++

	var id int
	fmt.Sscanf(orgGUID, "cf_org_id_%d", &id)
	//v2
	if m.CfClientVersion == "V2" {
		return cfclient.Org{
			Guid: orgGUID,
			Name: fmt.Sprintf("cf_org_name_%d", id),
		}, nil
	}
	//v3
	return &resource.Organization{
		Name:     fmt.Sprintf("cf_org_name_%d", id),
		Resource: resource.Resource{GUID: orgGUID}}, nil
}

func (m *AppClientMock) CreateApp(appID, spaceID string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	//v2
	if m.CfClientVersion == "V2" {
		app := cfclient.App{
			Guid:      appID,
			Name:      appID,
			SpaceGuid: spaceID,
		}
		m.apps[appID] = app
	} else {
		app := resource.App{
			Resource:      resource.Resource{GUID: appID},
			Name:          appID,
			Relationships: resource.SpaceRelationship{Space: resource.ToOneRelationship{Data: &resource.Relationship{GUID: spaceID}}},
			Metadata:      &resource.Metadata{Labels: make(map[string]*string)},
		}
		m.apps[appID] = &app
	}
	//v3

	//v2
	//m.apps[appID] = app
	//v3
	//m.apps[appID] = &app
}

func (m *AppClientMock) DeleteApp(appID string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.apps, appID)
}

// v2
//func getApps(n int) map[string]cfclient.App {
//	apps := make(map[string]cfclient.App, n)
//	for i := 0; i < n; i++ {
//		app := cfclient.App{
//			Guid:      fmt.Sprintf("cf_app_id_%d", i),
//			Name:      fmt.Sprintf("cf_app_name_%d", i),
//			SpaceGuid: fmt.Sprintf("cf_space_id_%d", i%50),
//		}
//		apps[app.Guid] = app
//	}
//	return apps
//}

// v3
func getApps(n int, cfclientversion string) map[string]*resource.App {

	apps := make(map[string]*resource.App, n)
	for i := 0; i < n; i++ {
		if cfclientversion == "V2" {
			app := resource.App{
				Guid:      fmt.Sprintf("cf_app_id_%d", i),
				Name:      fmt.Sprintf("cf_app_name_%d", i),
				SpaceGuid: fmt.Sprintf("cf_space_id_%d", i%50),
			}
			apps[app.Guid] = app
		} else {
			app := resource.App{
				Resource:      resource.Resource{GUID: fmt.Sprintf("cf_app_id_%d", i)},
				Name:          fmt.Sprintf("cf_app_name_%d", i),
				Relationships: resource.SpaceRelationship{Space: resource.ToOneRelationship{Data: &resource.Relationship{GUID: fmt.Sprintf("cf_space_id_%d", i%50)}}},
				Metadata:      &resource.Metadata{Labels: make(map[string]*string)},
			}
			apps[app.GUID] = &app
		}
	}
	return apps
}

func (m *AppClientMock) ListAppsCallCount() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.listAppsCallCount
}

func (m *AppClientMock) AppByGUIDCallCount() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.appByGUIDCallCount
}

func (m *AppClientMock) GetOrgByGUIDCallCount() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.getOrgByGUIDCallCount
}

func (m *AppClientMock) GetSpaceByGUIDCallCount() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.getSpaceByGUIDCallCount
}

func (m *AppClientMock) ResetCallCounts() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.listAppsCallCount = 0
	m.appByGUIDCallCount = 0
	m.getOrgByGUIDCallCount = 0
	m.getSpaceByGUIDCallCount = 0
}
