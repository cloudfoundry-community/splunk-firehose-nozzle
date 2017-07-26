package drain

const (
	CF_ORG_ID   = "cf_org_id"
	CF_SPACE_ID = "cf_space_id"
	CF_APP_ID   = "cf_app_id"

	CF_APP_NAME   = "cf_app_name"
	CF_SPACE_NAME = "cf_space_name"
	CF_ORG_NAME   = "cf_org_name"
)

type IndexMap struct {
	By    string  `json:"by" binding:"required"`
	Value string  `json:"value" binding:"required"`
	Index *string `json:"index" binding:"required"`
}

// {
//    "defult_index": "main",
//    "mappings": [
//	    {"by": "cf_app_id", "value": "app uuid", "index": "sales"},
//	    {"by": "cf_space_id", "value": "space uuid", index": "test"},
//	    {"by": "cf_org_id", "value": "org uuid", "index": "financial"}
//    ]
// }

type IndexMapConfig struct {
	DefaultIndex  string      `json:"default_index" binding:"required"`
	Mappings      []*IndexMap `json:"mappings" binding:"required"`
	removeAppMeta bool
}

// NeedsAppInfo determin if we need query app meta data information for index routing.
// If we need do index routing accordingly to space/org ID, we will need
// App meda data information.
func (c *IndexMapConfig) NeedsAppInfo(addAppInfo bool) bool {
	for _, mapping := range c.Mappings {
		if mapping.By == CF_ORG_ID || mapping.By == CF_SPACE_ID {
			if !addAppInfo {
				// If uses doesn't want index app info, we clean it up after
				// done with index routing
				c.removeAppMeta = true
			}
			return true
		}
	}
	return false
}

type IndexRouting struct {
	config *IndexMapConfig
}

// config param should be validated by clients by calling
// before calling NewIndexRouting
func NewIndexRouting(config *IndexMapConfig) *IndexRouting {
	return &IndexRouting{
		config: config,
	}
}

func (i *IndexRouting) removeAppMeta(fields map[string]interface{}) {
	if !i.config.removeAppMeta {
		return
	}

	keys := [...]string{CF_APP_NAME, CF_SPACE_NAME, CF_ORG_NAME, CF_SPACE_ID, CF_ORG_ID}
	for i := range keys {
		delete(fields, keys[i])
	}
}

// LookupIndex goes through the indexing rules one by one, once there is a batch
// is found, the index will be returned immediately. Otherwise a default Index will be
// returned.
func (i *IndexRouting) LookupIndex(fields map[string]interface{}) *string {
	idx := &i.config.DefaultIndex
	for _, idxMap := range i.config.Mappings {
		id, ok := fields[idxMap.By].(string)
		if ok && id == idxMap.Value {
			idx = idxMap.Index
			break
		}
	}

	i.removeAppMeta(fields)

	return idx
}
