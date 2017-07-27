package sink

import (
	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/logging"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/splunk"
	"net"
	"os"
)

type SplunkSink struct {
	name         string
	index        string
	host         string
	hostIP       string
	splunkClient splunk.SplunkClient
}

func NewSplunkSink(name string, index string, host string, splunkClient splunk.SplunkClient) *SplunkSink {
	hostname, hostIP := GetHostInfo(host)
	return &SplunkSink{
		name:         name,
		index:        index,
		host:         hostname,
		hostIP:       hostIP,
		splunkClient: splunkClient,
	}
}

func (s *SplunkSink) Log(message lager.LogFormat) {
	event := map[string]interface{}{
		"job_index":     s.index,
		"job":           s.name,
		"ip":            s.hostIP,
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

// get host details once to avoid costly os calls
func GetHostInfo(host string) (string, string) {
	var hostname string
	var err error
	if host == "" {
		hostname, err = os.Hostname()
		if err != nil {
			logging.LogError("Unable to get host name", err)
		}
	} else {
		hostname = host
	}

	ipAddresses, err := net.LookupIP(hostname)
	if err != nil {
		logging.LogError("Unable to get IP from host name", err)
	}

	for _, ia := range ipAddresses {
		return hostname, ia.String()
	}

	return hostname, hostname
}
