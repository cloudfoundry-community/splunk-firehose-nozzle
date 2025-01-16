package cache

import (
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"net/url"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

type App struct {
	Name        string
	Guid        string
	SpaceName   string
	SpaceGuid   string
	OrgName     string
	OrgGuid     string
	CfAppEnv    map[string]interface{} //V2
	CfAppLabels map[string]*string     //V3
	IgnoredApp  bool
}

type Cache interface {
	Open() error
	Close() error
	GetAllApps() (map[string]*App, error)
	GetApp(string) (*App, error)
}

type AppClient struct {
	version string
	v2      AppClientV2
	v3      AppClientV3
}

type AppClientV2 interface {
	AppByGuid(appGuid string) (cfclient.App, error)
	ListApps() ([]cfclient.App, error)
	ListAppsByQueryWithLimits(query url.Values, totalPages int) ([]cfclient.App, error)
	GetSpaceByGuid(spaceGUID string) (cfclient.Space, error)
	GetOrgByGuid(orgGUID string) (cfclient.Org, error)
}

type AppClientV3 interface {
	AppByGuid(appGuid string) (*resource.App, error)
	ListApps() ([]*resource.App, error)
	GetSpaceByGuid(spaceGUID string) (*resource.Space, error)
	GetOrgByGuid(orgGUID string) (*resource.Organization, error)
}
