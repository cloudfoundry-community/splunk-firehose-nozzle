package drain

import (
	"errors"
	"regexp"
)

const (
	CF_ORG_NAME   = "cf_org_name"
	CF_SPACE_NAME = "cf_space_name"
	CF_APP_NAME   = "cf_app_name"
)

type IndexMap struct {
	Value string  `json:"value" binding:"required"`
	Index *string `json:"index" binding:"required"`
	regex *regexp.Regexp
}

// {
//    "defult_index": "main",
//    "mappings": {
//	    "cf_org_name": [{"value": "sales.*", "index": "sales"}, ...],
//	    "cf_space_name": [{"value": "test.*", "index": "test"}, ...],
//	    "cf_app_name": [{"value": "fin.*", "index": "financial"}, {"value": "fin.*", "index": nil}, ...]
//	  }
// }

type IndexMapConfig struct {
	DefaultIndex string                 `json:"default_index" binding:"required"`
	Mappings     map[string][]*IndexMap `json:"mappings" binding:"required"`
}

// Validate validates if the index mapping configuration is valid
// and it will populate some internal states if the config is good
func (c *IndexMapConfig) Validate() error {
	indexConfiged := false

	for _, idxMaps := range c.Mappings {
		for _, idxMap := range idxMaps {
			if idxMap.Index != nil && *idxMap.Index != "" {
				indexConfiged = true
			}
			regex, err := regexp.Compile(idxMap.Value)
			if err != nil {
				return err
			}
			idxMap.regex = regex
		}
	}

	if !indexConfiged && c.DefaultIndex == "" {
		return errors.New("No index has been configured")
	}

	return nil
}

type IndexRouting struct {
	config *IndexMapConfig
}

// config param should be validated by clients by calling config.Validate())
// before calling NewIndexRouting
func NewIndexRouting(config *IndexMapConfig) *IndexRouting {
	return &IndexRouting{
		config: config,
	}
}

// LookupIndex finds a matching Splunk Index for "fields" passed according to the
// configuration provided by clients
// It first searches the index mapping configuration in the following order
// app -> space -> org
// If during the search, there is a match found, the corresponding index will be
// returned immediately. Otherwise a default Index will be returned
// Note: if this function returns a nil string, it means the "fields" (event) can be
// discarded directly
func (i *IndexRouting) LookupIndex(fields map[string]interface{}) *string {
	if i.config.Mappings == nil {
		return &i.config.DefaultIndex
	}

	// app has highest priority
	for _, byName := range []string{CF_APP_NAME, CF_SPACE_NAME, CF_ORG_NAME} {
		name, ok := fields[byName].(string)
		if !ok {
			continue
		}

		idxMaps := i.config.Mappings[byName]
		if idxMaps == nil {
			continue
		}

		for _, idxMap := range idxMaps {
			if idxMap.regex == nil {
				continue
			}

			if idxMap.regex.MatchString(name) {
				return idxMap.Index
			}
		}
	}

	return &i.config.DefaultIndex
}
