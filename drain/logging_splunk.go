package drain

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/config"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/splunk"
)

type LoggingSplunk struct {
	logger      lager.Logger
	client      splunk.SplunkClient
	flushWindow time.Duration

	defaultIndex  string
	indexMappings []config.Mapping

	batch []map[string]interface{}
}

func NewLoggingSplunk(logger lager.Logger, splunkClient splunk.SplunkClient, flushWindow time.Duration, defaultIndex string, indexMappings []config.Mapping) *LoggingSplunk {
	return &LoggingSplunk{
		logger:        logger,
		client:        splunkClient,
		flushWindow:   flushWindow,
		defaultIndex:  defaultIndex,
		indexMappings: indexMappings,
	}
}

func (l *LoggingSplunk) Connect() bool {
	go l.consume()

	return true
}

func (l *LoggingSplunk) ShipEvents(fields map[string]interface{}, msg string) {
	l.batch = append(l.batch, l.buildEvent(fields, msg))
}

func (l *LoggingSplunk) consume() {
	ticker := time.Tick(l.flushWindow)
	for range ticker {
		if len(l.batch) > 0 {
			l.logger.Info(fmt.Sprintf("Posting %d events", len(l.batch)))
			err := l.client.Post(l.batch)
			if err != nil {
				l.logger.Fatal("Unable to talk to Splunk", err)
			}

			l.batch = make([]map[string]interface{}, 0)
		} else {
			l.logger.Debug("No events to post")
		}
	}
}

func (l *LoggingSplunk) buildEvent(fields map[string]interface{}, msg string) map[string]interface{} {
	if len(msg) > 0 {
		fields["msg"] = msg
	}
	event := map[string]interface{}{}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	if val, ok := fields["timestamp"]; ok {
		timestamp = l.nanoSecondsToSeconds(val.(int64))
	}
	event["time"] = timestamp

	event["host"] = fields["ip"]
	event["source"] = fields["job"]

	eventType := strings.ToLower(fields["event_type"].(string))
	event["sourcetype"] = fmt.Sprintf("cf:%s", eventType)

	event["event"] = fields

	if l.defaultIndex != "" {
		event["index"] = l.defaultIndex
	}
	for _, mapping := range l.indexMappings {
		if val, ok := fields[mapping.Key]; ok {
			if val == mapping.Value {
				event["index"] = mapping.Index
			}
		}
		break
	}

	return event
}

func (l *LoggingSplunk) nanoSecondsToSeconds(nanoseconds int64) string {
	seconds := float64(nanoseconds) * math.Pow(1000, -3)
	return fmt.Sprintf("%.3f", seconds)
}
