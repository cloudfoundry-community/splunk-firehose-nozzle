package eventrouter

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/cache"
	fevents "github.com/cloudfoundry-community/splunk-firehose-nozzle/events"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/extrafields"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/logging"
	"github.com/cloudfoundry/sonde-go/events"
)

type Router interface {
	Route(msg *events.Envelope) error
	Setup(wantedEvents string) error
	SelectedEvents() map[string]bool
	SelectedEventsCount() map[string]uint64
	TotalCountOfSelectedEvents() uint64
	LogEventTotals(logTotalsTime time.Duration)
}

func IsAuthorizedEvent(wantedEvent string) bool {
	for _, authorizeEvent := range events.Envelope_EventType_name {
		if wantedEvent == authorizeEvent {
			return true
		}
	}
	return false
}

func GetListAuthorizedEventEvents() (authorizedEvents string) {
	arrEvents := []string{}
	for _, listEvent := range events.Envelope_EventType_name {
		arrEvents = append(arrEvents, listEvent)
	}
	sort.Strings(arrEvents)
	return strings.Join(arrEvents, ", ")
}

type router struct {
	appCache            cache.Cache
	selectedEvents      map[string]bool
	selectedEventsCount map[string]uint64
	mutex               *sync.Mutex
	extraFields         map[string]string
	log                 logging.Logging
}

func New(appCache cache.Cache, log logging.Logging) Router {
	return &router{
		appCache:            appCache,
		selectedEvents:      make(map[string]bool),
		selectedEventsCount: make(map[string]uint64),
		log:                 log,
		mutex:               &sync.Mutex{},
		extraFields:         make(map[string]string),
	}
}

func (r *router) SelectedEvents() map[string]bool {
	return r.selectedEvents
}

func (r *router) Route(msg *events.Envelope) error {
	eventType := msg.GetEventType()

	if r.selectedEvents[eventType.String()] {
		var event *fevents.Event
		switch eventType {
		case events.Envelope_HttpStartStop:
			event = fevents.HttpStartStop(msg)
		case events.Envelope_LogMessage:
			event = fevents.LogMessage(msg)
		case events.Envelope_ValueMetric:
			event = fevents.ValueMetric(msg)
		case events.Envelope_CounterEvent:
			event = fevents.CounterEvent(msg)
		case events.Envelope_Error:
			event = fevents.ErrorEvent(msg)
		case events.Envelope_ContainerMetric:
			event = fevents.ContainerMetric(msg)
		}

		event.AnnotateWithEnveloppeData(msg)

		event.AnnotateWithMetaData(r.extraFields)
		if _, hasAppId := event.Fields["cf_app_id"]; hasAppId {
			event.AnnotateWithAppData(r.appCache)
		}

		r.mutex.Lock()
		//We do not ship Event
		if ignored, hasIgnoredField := event.Fields["cf_ignored_app"]; ignored == true && hasIgnoredField {
			r.selectedEventsCount["ignored_app_message"]++
		} else {
			err := r.log.Log(event.Fields, event.Msg)
			if err != nil {
				fields := map[string]interface{}{"err": fmt.Sprintf("%s", err)}
				r.log.Log(fields, "Failed to ship events")
			}
			r.selectedEventsCount[eventType.String()]++

		}
		r.mutex.Unlock()
	}
	return nil
}

func (r *router) Setup(wantedEvents string) error {
	r.selectedEvents = make(map[string]bool)

	if wantedEvents == "" {
		r.selectedEvents["LogMessage"] = true
		return nil
	}

	for _, event := range strings.Split(wantedEvents, ",") {
		event = strings.TrimSpace(event)
		if IsAuthorizedEvent(event) {
			r.selectedEvents[event] = true
		} else {
			return fmt.Errorf("Rejected event name [%s] - Valid events: %s", event, GetListAuthorizedEventEvents())
		}
	}
	return nil
}

func (r *router) SetExtraFields(extraEventsString string) error {
	extraFields, err := extrafields.ParseExtraFields(extraEventsString)
	if err != nil {
		return err
	}
	r.extraFields = extraFields
	return nil
}

func (r *router) TotalCountOfSelectedEvents() uint64 {
	var total = uint64(0)
	for _, count := range r.SelectedEventsCount() {
		total += count
	}
	return total
}

func (r *router) SelectedEventsCount() map[string]uint64 {
	return r.selectedEventsCount
}

func (r *router) LogEventTotals(logTotalsTime time.Duration) {
	firehoseEventTotals := time.NewTicker(logTotalsTime)
	count := uint64(0)
	startTime := time.Now()
	totalTime := startTime

	go func() {
		for range firehoseEventTotals.C {
			elapsedTime := time.Since(startTime).Seconds()
			totalElapsedTime := time.Since(totalTime).Seconds()
			startTime = time.Now()
			event, lastCount := r.getEventTotals(totalElapsedTime, elapsedTime, count)
			count = lastCount
			r.log.Log(event.Fields, event.Msg)
		}
	}()
}

func (r *router) getEventTotals(totalElapsedTime float64, elapsedTime float64, lastCount uint64) (*fevents.Event, uint64) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	totalCount := r.TotalCountOfSelectedEvents()
	sinceLastTime := float64(int(elapsedTime*10)) / 10
	fields := logrus.Fields{
		"total_count":   totalCount,
		"by_sec_Events": int((totalCount - lastCount) / uint64(sinceLastTime)),
	}

	for eventtype, count := range r.SelectedEventsCount() {
		fields[eventtype] = count
	}

	event := &fevents.Event{
		Type:   "firehose_to_syslog_stats",
		Msg:    "Statistic for firehose to syslog",
		Fields: fields,
	}
	event.AnnotateWithMetaData(map[string]string{})
	return event, totalCount
}
