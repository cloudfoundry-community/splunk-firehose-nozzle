package logging

import (
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
)

type LoggingLogrus struct {
	Logger           *logrus.Logger
	syslogServer     string
	debugFlag        bool
	logFormatterType string
	certPath         string
	syslogProtocol   string
}

func NewLogging(SyslogServerFlag string, SysLogProtocolFlag string, LogFormatterFlag string, certP string, DebugFlag bool) Logging {
	return &LoggingLogrus{
		Logger:           logrus.New(),
		syslogServer:     SyslogServerFlag,
		logFormatterType: LogFormatterFlag,
		syslogProtocol:   SysLogProtocolFlag,
		certPath:         certP,
		debugFlag:        DebugFlag,
	}
}

func (l *LoggingLogrus) Connect() bool {

	success := false
	l.Logger.Formatter = GetLogFormatter(l.logFormatterType)

	if !l.debugFlag {
		l.Logger.Out = ioutil.Discard
	} else {
		l.Logger.Out = os.Stdout
	}

	return success
}

func (l *LoggingLogrus) ShipEvents(eventFields map[string]interface{}, Message string) {
	l.Logger.WithFields(eventFields).Info(Message)
}

func GetLogFormatter(logFormatterType string) logrus.Formatter {
	switch logFormatterType {
	case "text":
		return &logrus.TextFormatter{}
	default:
		return &logrus.JSONFormatter{}
	}
}
