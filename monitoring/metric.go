package monitoring

import (
	"strings"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventwriter"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/utils"
)

const (
	SplunkMetric = "metric_name:"
)

type Metrics struct {
	CallerFuncs   map[string]MonitorFunc
	Counters      map[string][]utils.Counter
	logger        lager.Logger
	interval      time.Duration
	ticker        *time.Ticker
	writer        eventwriter.Writer
	filtermetrics string
}

var (
	monitor Monitor = &NoMonitor{}
)

func InitMetrics(logger lager.Logger, interval time.Duration, writer eventwriter.Writer, filter string) *Metrics {
	// initMetricOnce.Do(func() {
	if monitor != nil {
		monitor.Stop()
	}
	monitor = &Metrics{
		CallerFuncs:   make(map[string]MonitorFunc),
		Counters:      make(map[string][]utils.Counter),
		logger:        logger,
		interval:      interval,
		writer:        writer,
		filtermetrics: filter,
	}
	return monitor.(*Metrics)
}

func (m *Metrics) RegisterFunc(id string, mFunc MonitorFunc) {
	listOfMetrics := strings.Split(m.filtermetrics, ",")
	if contains(listOfMetrics, id) && m.interval > 0*time.Second {
		m.CallerFuncs[id] = mFunc
	}
}

func RegisterFunc(id string, callerFunc MonitorFunc) {
	monitor.RegisterFunc(id, callerFunc)
}

func (m *Metrics) RegisterCounter(id string, varType utils.CounterType) utils.Counter {

	listOfMetrics := strings.Split(m.filtermetrics, ",")
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

func (m *Metrics) ExtractFunc(metricEvent map[string]interface{}) {
	for key, Func := range m.CallerFuncs {
		valofFunc := Func()
		splunkKey := SplunkMetric + key
		metricEvent[splunkKey] = valofFunc
	}
}

func (m *Metrics) ExtractCounter(metricEvent map[string]interface{}) {
	for key, listofCounters := range m.Counters {
		InitSum := listofCounters[0]
		for i := 1; i < len(listofCounters); i++ {
			InitSum.Add(listofCounters[i].Value())
			listofCounters[i].Reset()
		}
		splunkKey := SplunkMetric + key
		metricEvent[splunkKey] = InitSum.Value()
		listofCounters[0].Reset()
	}
}

func (m *Metrics) Start() error {
	if m.filtermetrics != "" {
		ticker := time.NewTicker(m.interval)
		m.ticker = ticker

		metricEvent := make(map[string]interface{})
		for {
			select {
			case <-ticker.C:
				m.ExtractFunc(metricEvent)
				m.ExtractCounter(metricEvent)
				events := []map[string]interface{}{
					metricEvent,
				}
				m.writer.Write(events)
			}
		}

	}
	return nil

}

func (m *Metrics) Stop() error {
	m.ticker.Stop()
	return nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
