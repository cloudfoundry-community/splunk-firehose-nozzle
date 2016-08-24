package splunk

import (
	"encoding/json"
	"code.cloudfoundry.org/lager"
)

type LoggingSplunk struct {
	Logger       lager.Logger
	SplunkClient SplunkClient
}

func NewLoggingSplunk(logger lager.Logger, splunkClient SplunkClient) *LoggingSplunk {
	return &LoggingSplunk{
		Logger:       logger,
		SplunkClient: splunkClient,
	}
}

func (l *LoggingSplunk) Connect() bool {
	return true
}

func (l *LoggingSplunk) ShipEvents(fields map[string]interface{}, msg string) {
	if len(msg) > 0 {
		fields["msg"] = msg
	}

	eventJson, err := json.Marshal(fields)
	if err != nil {
		l.Logger.Error("Unable to marshall event json", err)
	} else {
		//todo: batching
		l.SplunkClient.Post([]*[]byte{&eventJson})
	}
}
