package events

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/cache"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/utils"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/sirupsen/logrus"
)

type Event struct {
	Fields map[string]interface{}
	Msg    string
	Type   string
}

type Config struct {
	SelectedEvents string
	AddAppName     bool
	AddOrgName     bool
	AddOrgGuid     bool
	AddSpaceName   bool
	AddSpaceGuid   bool
	AddTags        bool
}

var AppMetadata = []string{
	"AppName",
	"OrgName",
	"OrgGuid",
	"SpaceName",
	"SpaceGuid",
}

func HttpStart(msg *events.Envelope) *Event {
	httpStart := msg.GetHttpStart()
	fields := logrus.Fields{
		"timestamp":         httpStart.GetTimestamp(),
		"request_id":        utils.FormatUUID(httpStart.GetRequestId()),
		"method":            httpStart.GetMethod().String(),
		"uri":               httpStart.GetUri(),
		"remote_addr":       httpStart.GetRemoteAddress(),
		"user_agent":        httpStart.GetUserAgent(),
		"parent_request_id": utils.FormatUUID(httpStart.GetParentRequestId()),
		"cf_app_id":         utils.FormatUUID(httpStart.GetApplicationId()),
		"instance_index":    httpStart.GetInstanceIndex(),
		"instance_id":       httpStart.GetInstanceId(),
	}

	return &Event{
		Fields: fields,
		Msg:    "",
	}
}

func HttpStop(msg *events.Envelope) *Event {
	httpStop := msg.GetHttpStop()

	fields := logrus.Fields{
		"timestamp":      httpStop.GetTimestamp(),
		"uri":            httpStop.GetUri(),
		"request_id":     utils.FormatUUID(httpStop.GetRequestId()),
		"peer_type":      httpStop.GetPeerType().String(),
		"status_code":    httpStop.GetStatusCode(),
		"content_length": httpStop.GetContentLength(),
		"cf_app_id":      utils.FormatUUID(httpStop.GetApplicationId()),
	}

	return &Event{
		Fields: fields,
		Msg:    "",
	}
}

func HttpStartStop(msg *events.Envelope) *Event {
	httpStartStop := msg.GetHttpStartStop()

	fields := logrus.Fields{
		"cf_app_id":       utils.FormatUUID(httpStartStop.GetApplicationId()),
		"content_length":  httpStartStop.GetContentLength(),
		"instance_id":     httpStartStop.GetInstanceId(),
		"instance_index":  httpStartStop.GetInstanceIndex(),
		"method":          httpStartStop.GetMethod().String(),
		"peer_type":       httpStartStop.GetPeerType().String(),
		"remote_addr":     httpStartStop.GetRemoteAddress(),
		"request_id":      utils.FormatUUID(httpStartStop.GetRequestId()),
		"start_timestamp": httpStartStop.GetStartTimestamp(),
		"status_code":     httpStartStop.GetStatusCode(),
		"stop_timestamp":  httpStartStop.GetStopTimestamp(),
		"uri":             httpStartStop.GetUri(),
		"user_agent":      httpStartStop.GetUserAgent(),
		"duration_ms":     (((httpStartStop.GetStopTimestamp() - httpStartStop.GetStartTimestamp()) / 1000) / 1000),
		"forwarded":       httpStartStop.GetForwarded(),
	}

	return &Event{
		Fields: fields,
		Msg:    "",
	}
}

func LogMessage(msg *events.Envelope) *Event {
	logMessage := msg.GetLogMessage()

	fields := logrus.Fields{
		"cf_app_id":       logMessage.GetAppId(),
		"timestamp":       logMessage.GetTimestamp(),
		"source_type":     logMessage.GetSourceType(),
		"message_type":    logMessage.GetMessageType().String(),
		"source_instance": logMessage.GetSourceInstance(),
	}

	return &Event{
		Fields: fields,
		Msg:    string(logMessage.GetMessage()),
	}
}

func ValueMetric(msg *events.Envelope) *Event {
	valMetric := msg.GetValueMetric()
	value := valMetric.GetValue()

	fields := logrus.Fields{
		"name":  valMetric.GetName(),
		"unit":  valMetric.GetUnit(),
		"value": value,
	}

	// Convert special values
	if math.IsNaN(value) {
		fields["value"] = "NaN"
	} else if math.IsInf(value, 1) {
		fields["value"] = "Infinity"
	} else if math.IsInf(value, -1) {
		fields["value"] = "-Infinity"
	}

	return &Event{
		Fields: fields,
		Msg:    "",
	}
}

func CounterEvent(msg *events.Envelope) *Event {
	counterEvent := msg.GetCounterEvent()

	fields := logrus.Fields{
		"name":  counterEvent.GetName(),
		"delta": counterEvent.GetDelta(),
		"total": counterEvent.GetTotal(),
	}

	return &Event{
		Fields: fields,
		Msg:    "",
	}
}

func ErrorEvent(msg *events.Envelope) *Event {
	errorEvent := msg.GetError()

	fields := logrus.Fields{
		"code":   errorEvent.GetCode(),
		"source": errorEvent.GetSource(),
	}

	return &Event{
		Fields: fields,
		Msg:    errorEvent.GetMessage(),
	}
}

func ContainerMetric(msg *events.Envelope) *Event {
	containerMetric := msg.GetContainerMetric()

	fields := logrus.Fields{
		"cf_app_id":          containerMetric.GetApplicationId(),
		"cpu_percentage":     containerMetric.GetCpuPercentage(),
		"disk_bytes":         containerMetric.GetDiskBytes(),
		"disk_bytes_quota":   containerMetric.GetDiskBytesQuota(),
		"instance_index":     containerMetric.GetInstanceIndex(),
		"memory_bytes":       containerMetric.GetMemoryBytes(),
		"memory_bytes_quota": containerMetric.GetMemoryBytesQuota(),
	}

	return &Event{
		Fields: fields,
		Msg:    "",
	}
}

func (e *Event) AnnotateWithAppData(appCache cache.Cache, config *Config) {
	cf_app_id := e.Fields["cf_app_id"]
	appGuid := fmt.Sprintf("%s", cf_app_id)

	if cf_app_id != nil && appGuid != "<nil>" && cf_app_id != "" {
		appInfo, err := appCache.GetApp(appGuid)
		if err != nil {
			if err == cache.ErrMissingAndIgnored {
				logrus.Info(err.Error(), cf_app_id)
			} else {
				logrus.Error("Failed to fetch application metadata from remote: ", err)
			}
			return
		} else if appInfo == nil {
			return
		}
		cf_app_name := appInfo.Name
		cf_space_id := appInfo.SpaceGuid
		cf_space_name := appInfo.SpaceName
		cf_org_id := appInfo.OrgGuid
		cf_org_name := appInfo.OrgName
		cf_ignored_app := appInfo.IgnoredApp
		app_env := appInfo.CfAppEnv

		if cf_app_name != "" && config.AddAppName {
			e.Fields["cf_app_name"] = cf_app_name
		}

		if cf_space_id != "" && config.AddSpaceGuid {
			e.Fields["cf_space_id"] = cf_space_id
		}

		if cf_space_name != "" && config.AddSpaceName {
			e.Fields["cf_space_name"] = cf_space_name
		}

		if cf_org_id != "" && config.AddOrgGuid {
			e.Fields["cf_org_id"] = cf_org_id
		}

		if cf_org_name != "" && config.AddOrgName {
			e.Fields["cf_org_name"] = cf_org_name
		}

		if app_env["SPLUNK_INDEX"] != nil {
			e.Fields["info_splunk_index"] = app_env["SPLUNK_INDEX"]
		}

		if cf_ignored_app != false {
			e.Fields["cf_ignored_app"] = cf_ignored_app
		}

	}
}

func (e *Event) AnnotateWithCFMetaData() {
	e.Fields["event_type"] = e.Type
}

func (e *Event) AnnotateWithEnvelopeData(msg *events.Envelope, config *Config) {
	e.Fields["origin"] = msg.GetOrigin()
	e.Fields["deployment"] = msg.GetDeployment()
	e.Fields["ip"] = msg.GetIp()
	e.Fields["job"] = msg.GetJob()
	e.Fields["job_index"] = msg.GetIndex()
	e.Type = msg.GetEventType().String()

	if config.AddTags {
		e.Fields["tags"] = msg.GetTags()
	}
}

func IsAuthorizedEvent(wantedEvent string) bool {
	_, ok := events.Envelope_EventType_value[wantedEvent]
	return ok
}

func AuthorizedEvents() string { // nosemgrep false-positive : `Envelope_EventType_name` is not pointer.
	arrEvents := []string{}
	for _, listEvent := range events.Envelope_EventType_name {
		arrEvents = append(arrEvents, listEvent)
	}
	sort.Strings(arrEvents)
	return strings.Join(arrEvents, ", ")
}

func ParseSelectedEvents(wantedEvents string) (map[string]bool, error) {
	wantedEvents = strings.TrimSpace(wantedEvents)
	selectedEvents := make(map[string]bool)
	if wantedEvents == "" {
		selectedEvents["LogMessage"] = true
		return selectedEvents, nil
	}

	var events []string
	if err := json.Unmarshal([]byte(wantedEvents), &events); err != nil {
		events = strings.Split(wantedEvents, ",")
	}

	for _, event := range events {
		event = strings.TrimSpace(event)
		if IsAuthorizedEvent(event) {
			selectedEvents[event] = true
		} else {
			return nil, fmt.Errorf("rejected event name [%s] - valid events: %s", event, AuthorizedEvents())
		}
	}
	return selectedEvents, nil
}

func getKeyValueFromString(kvPair string) (string, string, error) {
	values := strings.Split(kvPair, ":")
	if len(values) != 2 {
		return "", "", fmt.Errorf("When splitting %s by ':' there must be exactly 2 values, got these values %s", kvPair, values)
	}
	return strings.TrimSpace(values[0]), strings.TrimSpace(values[1]), nil
}

func ParseExtraFields(extraEventsString string) (map[string]string, error) {
	extraEvents := map[string]string{}

	for _, kvPair := range strings.Split(extraEventsString, ",") {
		if kvPair != "" {
			cleaned := strings.TrimSpace(kvPair)
			k, v, err := getKeyValueFromString(cleaned)
			if err != nil {
				return nil, err
			}
			extraEvents[k] = v
		}
	}
	return extraEvents, nil
}

func AuthorizedMetadata() string {
	return strings.Join(AppMetadata, ", ")
}
