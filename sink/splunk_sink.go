package sink

import (
	"fmt"
	"net"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/splunk"
)

type SplunkSink struct {
	name         string
	index        string
	host         string
	hostIP       string
	splunkClient splunk.SplunkClient
}

func NewSplunkSink(name string, index string, host string, splunkClient splunk.SplunkClient) *SplunkSink {
	hostname, hostIP, err := GetHostIPInfo(host)
	if err != nil {
		event := map[string]interface{}{
			"host":  hostname,
			"index": index,
			"event": fmt.Sprintf("Failed to resolve hostname and IP, error=%s", err),
		}

		events := []map[string]interface{}{event}
		splunkClient.Post(events)
	}

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

// GetHostIPInfo returns hostname and corresponding IP address
// If empty host is passed in, the current hostname and IP of the host will be
// returned. If the IP of the hostname can't be resolved, an empty IP and an
// error will be returned
func GetHostIPInfo(host string) (string, string, error) {
	var hostname string
	var err error

	hostname = host
	if hostname == "" {
		hostname, err = os.Hostname()
		if err != nil {
			return host, "", err
		}
	}

	ipAddresses, err := net.LookupIP(hostname)
	if err != nil {
		return hostname, "", err
	}

	for _, ia := range ipAddresses {
		return hostname, ia.String(), nil
	}

	return hostname, "", nil
}
