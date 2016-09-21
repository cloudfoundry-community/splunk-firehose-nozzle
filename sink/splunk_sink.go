package sink

import (
	"code.cloudfoundry.org/lager"

	"github.com/cf-platform-eng/splunk-firehose-nozzle/splunk"
)

type SplunkSink struct {
	name         string
	index        int
	host         string
	splunkClient splunk.SplunkClient
}

func NewSplunkSink(name string, index int, host string, splunkClient splunk.SplunkClient) *SplunkSink {
	return &SplunkSink{
		name:         name,
		index:        index,
		host:         host,
		splunkClient: splunkClient,
	}
}

func (s *SplunkSink) Log(message lager.LogFormat) {
	event := map[string]interface{}{
		"index":         s.index,
		"job":           s.name,
		"ip":            s.host,
		"origin":        "splunk_nozzle",
		"logger_source": message.Source,
		"message":       message.Message,
		"log_level":     int(message.LogLevel),
	}
	if message.Data != nil && len(message.Data) > 0 {
		data := map[string]interface{}{}
		for key, value := range message.Data {
			data[key] = value
		}
		event["data"] = data
	}

	events := []map[string]interface{}{
		map[string]interface{}{
			"time":       message.Timestamp,
			"host":       s.host,
			"source":     s.name,
			"sourcetype": "cf:splunknozzle",
			"event":      event,
		},
	}

	s.splunkClient.Post(events)
}
