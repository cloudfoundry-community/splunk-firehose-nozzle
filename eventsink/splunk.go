package eventsink

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"sync/atomic"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventwriter"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/utils"
)

const SPLUNK_HEC_FIELDS_SUPPORT_VERSION = "6.4"

type SplunkConfig struct {
	FlushInterval  time.Duration
	QueueSize      int // consumer queue buffer size
	BatchSize      int
	Retries        int // No of retries to post events to HEC before dropping events
	Hostname       string
	Version        string
	SubscriptionID string
	ExtraFields    map[string]string
	TraceLogging   bool
	UUID           string

	Logger lager.Logger
}

type Splunk struct {
	writers    []eventwriter.Writer
	config     *SplunkConfig
	events     chan map[string]interface{}
	wg         sync.WaitGroup
	eventCount uint64

	// cached IP
	ip string
}

func NewSplunk(writers []eventwriter.Writer, config *SplunkConfig) *Splunk {
	hostname, ip, _ := utils.GetHostIPInfo(config.Hostname)
	config.Hostname = hostname

	return &Splunk{
		writers:    writers,
		config:     config,
		events:     make(chan map[string]interface{}, config.QueueSize),
		ip:         ip,
		eventCount: 0,
	}
}

func (s *Splunk) Open() error {
	for _, client := range s.writers[:len(s.writers)-1] {
		s.wg.Add(1)
		go s.consume(client)
	}
	return nil
}

func (s *Splunk) Close() error {
	// Notify the consume loop to drain events and exit
	close(s.events)
	s.wg.Wait()
	return nil
}

func (s *Splunk) Write(fields map[string]interface{}, msg string) error {
	if len(msg) > 0 {
		fields["msg"] = msg
	}
	s.events <- fields
	return nil
}

func (s *Splunk) consume(writer eventwriter.Writer) {
	defer s.wg.Done()

	var batch []map[string]interface{}
	timer := time.NewTimer(s.config.FlushInterval)

	// Flush takes place when 1) batch limit is reached. 2) flush window expires
LOOP:
	for {
		select {
		case event, ok := <-s.events:
			if !ok {
				// events chan has closed and we have drained all events in it
				break LOOP
			}

			event = s.buildEvent(event)
			batch = append(batch, event)
			if len(batch) >= s.config.BatchSize {
				batch = s.indexEvents(writer, batch)
				timer.Reset(s.config.FlushInterval) // reset channel timer
			}

		case <-timer.C:
			batch = s.indexEvents(writer, batch)
			timer.Reset(s.config.FlushInterval)
		}

	}
	// Last batch
	s.indexEvents(writer, batch)
}

// indexEvents indexes events to Splunk
// return nil when successful which clears all outstanding events
// return what the batch has if there is an error for next retry cycle
func (s *Splunk) indexEvents(writer eventwriter.Writer, batch []map[string]interface{}) []map[string]interface{} {
	if len(batch) == 0 {
		return batch
	}
	var err error
	for i := 0; i < s.config.Retries; i++ {
		err = writer.Write(batch)
		if err == nil {
			return nil
		}
		s.config.Logger.Error("Unable to talk to Splunk", err)
		time.Sleep(getRetryInterval(i))
	}
	s.config.Logger.Error("Finish retrying and dropping events", err, lager.Data{"events": len(batch)})
	return nil
}

func (s *Splunk) buildEvent(fields map[string]interface{}) map[string]interface{} {
	if msg, ok := fields["msg"]; ok {
		if msgStr, ok := msg.(string); ok && len(msgStr) > 0 {
			fields["msg"] = utils.ToJson(msgStr)
		}
	}

	event := map[string]interface{}{}

	var timestamp string
	if val, ok := fields["timestamp"]; ok {
		if v, ok := val.(int64); ok {
			timestamp = utils.NanoSecondsToSeconds(v)
		}
	}

	// Timestamp
	if timestamp == "" {
		timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	}
	event["time"] = timestamp

	event["host"] = fields["ip"]
	event["source"] = fields["job"]

	if eventType, ok := fields["event_type"].(string); ok {
		event["sourcetype"] = fmt.Sprintf("cf:%s", strings.ToLower(eventType))
	}

	extraFields := make(map[string]interface{})

	if s.config.TraceLogging {
		extraFields["nozzle-event-counter"] = strconv.FormatUint(atomic.AddUint64(&s.eventCount, 1), 10)
		extraFields["subscription-id"] = s.config.SubscriptionID
		extraFields["uuid"] = s.config.UUID
	}
	for k, v := range s.config.ExtraFields {
		extraFields[k] = v
	}

	if s.config.Version >= SPLUNK_HEC_FIELDS_SUPPORT_VERSION {
		event["fields"] = extraFields
	} else {
		fields["pcf-extra"] = extraFields
	}
	event["event"] = fields
	return event
}

// Log implements lager.Sink required interface
func (s *Splunk) Log(message lager.LogFormat) {
	e := map[string]interface{}{
		"logger_source": message.Source,
		"message":       message.Message,
		"ip":            s.ip,
		"origin":        "splunk_nozzle",
		"log_level":     int(message.LogLevel),
	}

	event := map[string]interface{}{
		"host":       s.config.Hostname,
		"sourcetype": "cf:splunknozzle",
		"event":      e,
	}

	if message.Timestamp != "" {
		event["time"] = message.Timestamp
	}

	if len(message.Data) > 0 {
		data := map[string]interface{}{}
		for key, value := range message.Data {
			data[key] = value
		}
		e["data"] = data
	}

	events := []map[string]interface{}{event}
	s.writers[len(s.writers)-1].Write(events)
}

func getRetryInterval(attempt int) time.Duration {
	// algorithm taken from https://en.wikipedia.org/wiki/Exponential_backoff
	timeInSec := 5 + (0.5 * (math.Exp2(float64(attempt)) - 1.0))
	return time.Millisecond * time.Duration(1000*timeInSec)
}
