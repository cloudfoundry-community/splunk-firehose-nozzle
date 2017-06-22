package drain

import (
	"code.cloudfoundry.org/lager"
	"fmt"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/splunk"
	"math"
	"strconv"
	"strings"
	"time"
)

type LoggingConfig struct {
	FlushInterval time.Duration
	QueueSize     int //consumer queue buffer size
	BatchSize     int
	Retries       int //No of retries to post events to HEC before dropping events
	MetadataMode  string //Add extra fields as metadata or in event payload
}

type LoggingSplunk struct {
	logger lager.Logger
	client splunk.SplunkClient
	config *LoggingConfig
	events chan map[string]interface{}
}

func NewLoggingSplunk(logger lager.Logger, splunkClient splunk.SplunkClient, config *LoggingConfig) *LoggingSplunk {
	return &LoggingSplunk{
		logger: logger,
		client: splunkClient,
		config: config,
		events: make(chan map[string]interface{}, config.QueueSize),
	}
}

func (l *LoggingSplunk) Connect() bool {
	go l.consume()

	return true
}

func (l *LoggingSplunk) ShipEvents(fields map[string]interface{}, msg string, ExtraFields map[string]interface{},) {
	event := l.buildEvent(fields, msg, ExtraFields)
	l.events <- event
}

func (l *LoggingSplunk) consume() {
	var batch []map[string]interface{}
	tickChan := time.NewTicker(l.config.FlushInterval).C

	// Either flush window or batch size reach limits, we flush
	for {
		select {
		case event := <-l.events:
			batch = append(batch, event)
			if len(batch) >= l.config.BatchSize {
				batch = l.indexEvents(batch)
			}
		case <-tickChan:
			batch = l.indexEvents(batch)
		}
	}
}

// indexEvents indexes events to Splunk
// return nil when sucessful which clears all outstanding events
// return what the batch has if there is an error for next retry cycle
func (l *LoggingSplunk) indexEvents(batch []map[string]interface{}) []map[string]interface{} {
	if len(batch) == 0 {
		return batch
	}
	var err error
	for i := 0; i < l.config.Retries; i++ {
		l.logger.Info(fmt.Sprintf("Posting %d events", len(batch)))
		err = l.client.Post(batch)
		if err == nil {
			return nil
		}
		l.logger.Error("Unable to talk to Splunk", err)
		time.Sleep(5 * time.Second)
	}
	l.logger.Error("Finish retrying and dropping events", err, lager.Data{"events": len(batch)})
	return nil
}

func (l *LoggingSplunk) buildEvent(fields map[string]interface{}, msg string, ExtraFields map[string]interface{}) map[string]interface{} {
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

	eventType := strings.ToLower(ExtraFields["event_type"].(string))
	event["sourcetype"] = fmt.Sprintf("cf:%s", eventType)


	if l.config.MetadataMode == "modern" {
		event["fields"] = ExtraFields
	} else {
		fields["pcf-metadata"] = ExtraFields
	}

	event["event"] = fields

	return event
}

func (l *LoggingSplunk) nanoSecondsToSeconds(nanoseconds int64) string {
	seconds := float64(nanoseconds) * math.Pow(1000, -3)
	return fmt.Sprintf("%.3f", seconds)
}
