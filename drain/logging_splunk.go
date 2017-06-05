package drain

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/splunk"
)

type LoggingSplunk struct {
	logger      lager.Logger
	client      splunk.SplunkClient
	flushWindow time.Duration
	events      chan map[string]interface{}
}

func NewLoggingSplunk(logger lager.Logger, splunkClient splunk.SplunkClient, flushWindow time.Duration) *LoggingSplunk {
	return &LoggingSplunk{
		logger:      logger,
		client:      splunkClient,
		flushWindow: flushWindow,
		// FIXME, make buffer size 100 configurable
		events: make(chan map[string]interface{}, 100),
	}
}

func (l *LoggingSplunk) Connect() bool {
	go l.consume()

	return true
}

func (l *LoggingSplunk) ShipEvents(fields map[string]interface{}, msg string) {
	event := l.buildEvent(fields, msg)
	l.events <- event
}

func (l *LoggingSplunk) consume() {
	var batch []map[string]interface{}
	// FIXME, make batchSize configurable
	batchSize := 50
	tickChan := time.NewTicker(l.flushWindow).C

	// Either flush window or batch size reach limits, we flush
	for {
		select {
		case event := <-l.events:
			batch = append(batch, event)
			if len(batch) >= batchSize {
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

	l.logger.Info(fmt.Sprintf("Posting %d events", len(batch)))
	err := l.client.Post(batch)
	if err != nil {
		l.logger.Error("Unable to talk to Splunk, error=%+v", err)
		// return back the batch for next retry
		return batch
	}

	return nil
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

	return event
}

func (l *LoggingSplunk) nanoSecondsToSeconds(nanoseconds int64) string {
	seconds := float64(nanoseconds) * math.Pow(1000, -3)
	return fmt.Sprintf("%.3f", seconds)
}
