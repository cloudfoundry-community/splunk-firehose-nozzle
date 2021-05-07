package drain

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/splunk"
)

type LoggingConfig struct {
	FlushInterval time.Duration
	QueueSize     int //consumer queue buffer size
	BatchSize     int
	Retries       int //No of retries to post events to HEC before dropping events

	AddAppInfoOrig bool
	IndexMapping   *IndexMapConfig
}

type LoggingSplunk struct {
	logger  lager.Logger
	clients []splunk.SplunkClient
	config  *LoggingConfig
	events  chan map[string]interface{}

	indexRouting *IndexRouting
}

func NewLoggingSplunk(logger lager.Logger, splunkClients []splunk.SplunkClient, config *LoggingConfig) *LoggingSplunk {
	return &LoggingSplunk{
		logger:  logger,
		clients: splunkClients,
		config:  config,
		events:  make(chan map[string]interface{}, config.QueueSize),

		indexRouting: NewIndexRouting(config.IndexMapping),
	}
}

func (l *LoggingSplunk) Connect() bool {
	for _, client := range l.clients {
		go l.consume(client)
	}

	return true
}

func (l *LoggingSplunk) ShipEvents(fields map[string]interface{}, msg string) {
	if len(msg) > 0 {
		fields["msg"] = msg
	}

	l.events <- fields
}

func (l *LoggingSplunk) consume(client splunk.SplunkClient) {
	var batch []map[string]interface{}
	tickChan := time.NewTicker(l.config.FlushInterval).C

	// Either flush window or batch size reach limits, we flush
	for {
		select {
		case fields := <-l.events:
			event := l.buildEvent(fields)
			if idx, ok := event["index"]; ok && idx == nil {
				// blacklist, drop it on the floor
				continue
			}

			batch = append(batch, event)
			if len(batch) >= l.config.BatchSize {
				batch = l.indexEvents(client, batch)
			}
		case <-tickChan:
			batch = l.indexEvents(client, batch)
		}
	}
}

// indexEvents indexes events to Splunk
// return nil when sucessful which clears all outstanding events
// return what the batch has if there is an error for next retry cycle
func (l *LoggingSplunk) indexEvents(client splunk.SplunkClient, batch []map[string]interface{}) []map[string]interface{} {
	if len(batch) == 0 {
		return batch
	}
	var err error
	for i := 0; i < l.config.Retries; i++ {
		// l.logger.Info(fmt.Sprintf("Posting %d events", len(batch)))
		err = client.Post(batch)
		if err == nil {
			return nil
		}
		l.logger.Error("Unable to talk to Splunk", err)
		time.Sleep(5 * time.Second)
	}
	l.logger.Error("Finish retrying and dropping events", err, lager.Data{"events": len(batch)})
	return nil
}

// ToJSON tries to detect the JSON pattern for msg first, if msg contains JSON pattern either
// a map or an array (for efficiency), it will try to convert msg to a JSON object. If the convertion
// success, a JSON object will be returned. Otherwise the original msg will be returned
// If the msg param doesn't contain the JSON pattern, the msg will be returned directly
func ToJson(msg string) interface{} {
	trimmed := strings.TrimSpace(msg)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		// Probably the msg can be converted to a map JSON object
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &m); err != nil {
			// Failed to convert to JSON object, just return the original msg
			return msg
		}
		return m
	} else if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		// Probably the msg can be converted to an array JSON object
		var a []interface{}
		if err := json.Unmarshal([]byte(trimmed), &a); err != nil {
			// Failed to convert to JSON object, just return the original msg
			return msg
		}
		return a
	}
	return msg
}

func (l *LoggingSplunk) buildEvent(fields map[string]interface{}) map[string]interface{} {
	msg, ok := fields["msg"]
	if ok {
		// try to convert to json object
		msgStr, ok := msg.(string)
		if ok && len(msgStr) > 0 {
			fields["msg"] = ToJson(msgStr)
		}
	}

	event := map[string]interface{}{}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	if val, ok := fields["timestamp"]; ok {
		timestamp = l.nanoSecondsToSeconds(val.(int64))
	}
	event["time"] = timestamp

	event["host"] = fields["ip"]
	event["source"] = fields["job"]
	l.lookupIndex(fields, event)

	eventType := strings.ToLower(fields["event_type"].(string))
	event["sourcetype"] = fmt.Sprintf("cf:%s", eventType)

	event["event"] = fields

	return event
}

func (l *LoggingSplunk) lookupIndex(fields, event map[string]interface{}) {
	index := l.indexRouting.LookupIndex(fields)

	// If index is "", we can't set empty index due to HEC will
	// disard events with index=""
	if index != nil && *index != "" || index == nil {
		event["index"] = index
	}
}

func (l *LoggingSplunk) nanoSecondsToSeconds(nanoseconds int64) string {
	seconds := float64(nanoseconds) * math.Pow(1000, -3)
	return fmt.Sprintf("%.3f", seconds)
}
