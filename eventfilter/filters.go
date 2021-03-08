package eventfilter

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/sonde-go/events"
)

const (
	filterSep       = ";"
	filterValuesSep = ":"
)

// getterFunc is a function that, given an Envelope, returns the things we care
// about, e.g. the Deployment, Job, ...
type getterFunc = func(msg *events.Envelope) string

// supportedGetters are all supported keys we can use for filters and the
// getters / functions that pull the respective data out of an envelope.
var supportedGetters = map[string]getterFunc{
	"deployment": func(msg *events.Envelope) string {
		return msg.GetDeployment()
	},
	"origin": func(msg *events.Envelope) string {
		return msg.GetOrigin()
	},
	"job": func(msg *events.Envelope) string {
		return msg.GetJob()
	},
}

// filterFunc gets called with data from the message envelope, pulled out by a
// getter, and compares it to the user provided value. If it returns true, this
// message should be accepted, else the message should be dropped.
type filterFunc = func(msgData, userInput string) bool

// supportedFilters are all supported filter names and
// the filters / functions that match run against the data from the message and
// compares it to the data provided by the user.
// E.g. when we have
//   - the filter func strings.Contains
//   - the getter, that gets the message's origin
//   - and the value 'foo'
// only messages with the origin 'foo' will be accepted by the filter
var supportedFilters = map[string]filterFunc{
	"mustContain":    strings.Contains,
	"mustNotContain": func(msgData, userInput string) bool { return !strings.Contains(msgData, userInput) },
}

// SupportedFilterKeys lists all supported filter keys. This is only used to
// signal the list of supported keys to users, e.g. for the usage text.
var SupportedFilterKeys = func() []string {
	keys := make([]string, 0, len(supportedGetters))
	for k := range supportedGetters {
		keys = append(keys, k)
	}

	return keys
}()

// SupportedFilters lists all supported filter names. This is only used to
// signal the list of supported filters to users, e.g. for the usage text.
var SupportedFilters = func() []string {
	keys := make([]string, 0, len(supportedFilters))
	for k := range supportedFilters {
		keys = append(keys, k)
	}

	return keys
}()

// Filters is something that can tell it's Length (the number of its configured
// filters) and can be used to check if an envelope is accepted or should be
// dropped/discarded.
type Filters interface {
	Accepts(*events.Envelope) bool
	Length() int
}

type filterRule struct {
	getter getterFunc
	filter filterFunc
	value  string
}

var (
	errInvalidFormat    = fmt.Errorf("format must be '%q:%q:<value>'", SupportedFilterKeys, SupportedFilters)
	errEmptyValue       = fmt.Errorf("filter value must not be empty")
	errInvaldFilter     = fmt.Errorf("filter key must be one of %q", SupportedFilterKeys)
	errInvalidFilterKey = fmt.Errorf("filter must be one of %q", SupportedFilters)
)

func parseFilterConfig(filters string) ([]filterRule, error) {
	rules := []filterRule{}

	for _, filterRaw := range strings.Split(filters, filterSep) {
		filter := strings.TrimSpace(filterRaw)

		if filter == "" {
			continue
		}

		tokens := strings.Split(filter, filterValuesSep)
		if len(tokens) != 3 {
			return []filterRule{}, fmt.Errorf("filter %q invalid: %s", filter, errInvalidFormat)
		}

		getterKey := strings.TrimSpace(strings.ToLower(tokens[0]))
		filterKey := strings.TrimSpace(tokens[1])

		var ok bool
		rule := filterRule{
			value: tokens[2],
		}

		if rule.value == "" {
			return []filterRule{}, fmt.Errorf("filter %q invalid: %s", filter, errEmptyValue)
		}

		rule.filter, ok = supportedFilters[filterKey]
		if !ok {
			return []filterRule{}, fmt.Errorf("filter %q invalid: %s", filter, errInvaldFilter)
		}

		rule.getter, ok = supportedGetters[getterKey]
		if !ok {
			return []filterRule{}, fmt.Errorf("filter %q invalid: %s", filter, errInvalidFilterKey)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

type filter func(*events.Envelope) bool

type filters []filter

func (ef *filters) Accepts(msg *events.Envelope) bool {
	for _, f := range *ef {
		if allow := f(msg); !allow {
			return false
		}
	}

	return true
}

func (ef *filters) Length() int {
	return len(*ef)
}

func (ef *filters) addFilter(valueGetter getterFunc, valueFilter filterFunc, value string) {
	*ef = append(*ef, func(msg *events.Envelope) bool {
		return valueFilter(valueGetter(msg), value)
	})
}

func New(filterList string) (Filters, error) {
	f := &filters{}

	rules, err := parseFilterConfig(filterList)
	if err != nil {
		return nil, err
	}

	for _, rule := range rules {
		f.addFilter(rule.getter, rule.filter, rule.value)
	}

	return f, nil
}
