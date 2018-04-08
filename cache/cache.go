package cache

import (
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
	Environment map[string]interface{}
	SysEnv      map[string]interface{}
	IgnoredApp  bool
}

type Cache interface {
	Open() error
	Close() error

	GetAllApps() (map[string]*App, error)
	GetApp(string) (*App, error)
}

type AppClient interface {
	AppByGuid(appGuid string) (cfclient.App, error)
	GetAppEnv(appGuid string) (cfclient.AppEnv, error)
	ListApps() ([]cfclient.App, error)
	ListAppsByQueryWithLimits(query url.Values, totalPages int) ([]cfclient.App, error)
}
