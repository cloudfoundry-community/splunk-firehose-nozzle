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

//LoggingSplunk
//logger: 	 Logging client as implemented in lager
//client:	 Splunk Client as implemented in splunk/splunk_client
//flushWindow: 	 time in nanoseconds representing flush window
//events: 	 channel implementation of map[strings] to store events for Splunk to index
type LoggingSplunk struct {
	logger      lager.Logger
	client      splunk.SplunkClient
	flushWindow time.Duration
	events      chan map[string]interface{}
}

//"constructor" for NewLoggingSplunk
func NewLoggingSplunk(logger lager.Logger, splunkClient splunk.SplunkClient, flushWindow time.Duration) *LoggingSplunk {
	return &LoggingSplunk{
		logger:      logger,
		client:      splunkClient,
		flushWindow: flushWindow,
		// FIXME, make buffer size 100 configurable
		events: make(chan map[string]interface{}, 100),
	}
}

//Connect implements "firehose-to-syslog.logging.logging" by calling LoggingSplunk.consume() on LoggingSplunk object
func (l *LoggingSplunk) Connect() bool {
	go l.consume()

	return true
}

//ShipEvents implements "firehose-to-syslog.logging.logging" by buildingEvents and sending them to LoggingSplunk.events
func (l *LoggingSplunk) ShipEvents(fields map[string]interface{}, msg string) {
	event := l.buildEvent(fields, msg)
	l.events <- event
}

//consume function will send events through to indexEvents method.
// trigger on 2 scenarios:
// 1: BatchSize count reached
// 2: flushWindow Ticker timer reached
func (l *LoggingSplunk) consume() {
	var batch []map[string]interface{}
	// FIXME, make batchSize configurable
	batchSize := 50
	tickChan := time.NewTicker(l.flushWindow).C
	lastIndex := time.Now()

	// Either flush window or batch size reach limits, we flush
	for {
		select {
		case event := <-l.events:
			batch = append(batch, event)
			if len(batch) >= batchSize {
				lastIndex = time.Now()
				l.logger.Info("Index Events triggered by Batch Limit")
				batch = l.indexEvents(batch)

			}
		case <-tickChan:
			//Checking if index has happened within 90% of the flush window. Allowing for a small delay
			if time.Now().Sub(lastIndex).Seconds() >= 0.9*l.flushWindow.Seconds() {
				lastIndex = time.Now()
				l.logger.Info("Index Events triggered by Flush Window Time Expiry")
				batch = l.indexEvents(batch)
			}
		}
	}
}

// indexEvents indexes events to Splunk
// return nil when successful which clears all outstanding events
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

// buildEvent constructs a splunk event from fields and msg parameter
// returns constructed event in event variable
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

// nanoSecondsToSeconds is a simple helper function to convert nanoSeconds to Seconds
func (l *LoggingSplunk) nanoSecondsToSeconds(nanoseconds int64) string {
	seconds := float64(nanoseconds) * math.Pow(1000, -3)
	return fmt.Sprintf("%.3f", seconds)
}
