package sink

import (
	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/splunk"
	"os"
	"net"
)

type SplunkSink struct {
	name         string
	index        string
	host         string
	splunkClient splunk.SplunkClient
}

func NewSplunkSink(name string, index string, host string, splunkClient splunk.SplunkClient) *SplunkSink {

	if host == "" {
		hostname, err := os.Hostname()
		if err != nil {
			panic(err)
		}
		host = hostname
	}
	return &SplunkSink{
		name:         name,
		index:        index,
		host:         host,
		splunkClient: splunkClient,
	}
}

func (s *SplunkSink) Log(message lager.LogFormat) {

	host_ip_address, err := net.LookupIP(s.host)
	if err != nil {
		// what to do here?
		panic(err)
	}
	event := map[string]interface{}{
		"job_index":     s.index,
		"job":           s.name,
		"ip":            host_ip_address[0].String(),
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
