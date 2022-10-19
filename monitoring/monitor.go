package monitoring

import (
	"strings"
	"time"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/utils"
)

type MonitorFunc func() interface{}

type Monitor interface {
	RegisterFunc(string, MonitorFunc)
	RegisterCounter(string, utils.CounterType) utils.Counter
	Start()
	Stop() error
}

var (
	monitor Monitor = &NoMonitor{}
)

func (m *Metrics) RegisterFunc(id string, mFunc MonitorFunc) {
	listOfMetrics := strings.Split(m.selectedMonitoringMetrics, ",")
	if contains(listOfMetrics, id) && m.interval > 0*time.Second {
		m.CallerFuncs[id] = mFunc
	}
}

func RegisterFunc(id string, callerFunc MonitorFunc) {
	monitor.RegisterFunc(id, callerFunc)
}

func (m *Metrics) RegisterCounter(id string, varType utils.CounterType) utils.Counter {

	listOfMetrics := strings.Split(m.selectedMonitoringMetrics, ",")
	if contains(listOfMetrics, id) && m.interval > 0*time.Second {
		if varType == utils.UintType {
			ctr := new(utils.IntCounter)
			m.Counters[id] = append(m.Counters[id], ctr)
			return ctr
		}
	}
	return &utils.NopCounter{}
}

func RegisterCounter(id string, varType utils.CounterType) utils.Counter {
	return monitor.RegisterCounter(id, varType)
}
