package eventrouter

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/caching"
	fevents "github.com/cloudfoundry-community/splunk-firehose-nozzle/events"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/extrafields"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/logging"
	"github.com/cloudfoundry/sonde-go/events"
)

type Router interface {
	Route(msg *events.Envelope) error
	Setup(wantedEvents string) error
	SetExtraFields(extraEventsString string)
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
	appCache            caching.Caching
	selectedEvents      map[string]bool
	selectedEventsCount map[string]uint64
	mutex               *sync.Mutex
	log                 logging.Logging
	ExtraFields         map[string]string
}

func New(appCache caching.Caching, logging logging.Logging) Router {
	return &router{
		appCache:            appCache,
		selectedEvents:      make(map[string]bool),
		selectedEventsCount: make(map[string]uint64),
		log:                 logging,
		mutex:               &sync.Mutex{},
		ExtraFields:         make(map[string]string),
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

		event.AnnotateWithMetaData(r.ExtraFields)
		if _, hasAppId := event.Fields["cf_app_id"]; hasAppId {
			event.AnnotateWithAppData(r.appCache)
		}

		r.mutex.Lock()
		//We do not ship Event
		if ignored, hasIgnoredField := event.Fields["cf_ignored_app"]; ignored == true && hasIgnoredField {
			r.selectedEventsCount["ignored_app_message"]++
		} else {
			err := r.log.ShipEvents(event.Fields, event.Msg)
			if err != nil {
				logging.LogError("failed to ship events", err)
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
	} else {
		for _, event := range strings.Split(wantedEvents, ",") {
			if IsAuthorizedEvent(strings.TrimSpace(event)) {
				r.selectedEvents[strings.TrimSpace(event)] = true
				logging.LogStd(fmt.Sprintf("Event Type [%s] is included in the fireshose!", event), false)
			} else {
				return fmt.Errorf("Rejected Event Name [%s] - Valid events: %s", event, GetListAuthorizedEventEvents())
			}
		}
	}
	return nil
}

func (r *router) SetExtraFields(extraEventsString string) {
	// Parse extra fields from cmd call
	extraFields, err := extrafields.ParseExtraFields(extraEventsString)
	if err != nil {
		logging.LogError("Error parsing extra fields: ", err)
		os.Exit(1)
	}
	r.ExtraFields = extraFields
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
			r.log.ShipEvents(event.Fields, event.Msg)
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
