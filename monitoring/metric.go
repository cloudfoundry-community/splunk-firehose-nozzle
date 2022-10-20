package monitoring

import (
	"strings"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventwriter"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/utils"
)

const (
	splunkMetric = "metric_name:"
)

type Metrics struct {
	CallerFuncs               map[string]MonitorFunc
	Counters                  map[string][]utils.Counter
	logger                    lager.Logger
	interval                  time.Duration
	ticker                    *time.Ticker
	writer                    eventwriter.Writer
	selectedMonitoringMetrics *utils.Set
}

func NewMetricsMonitor(logger lager.Logger, interval time.Duration, writer eventwriter.Writer, filter string) Monitor {

	if monitor != nil {
		monitor.Stop()
	}
	monitor = &Metrics{
		CallerFuncs:               make(map[string]MonitorFunc),
		Counters:                  make(map[string][]utils.Counter),
		logger:                    logger,
		interval:                  interval,
		writer:                    writer,
		selectedMonitoringMetrics: setValuesForSet(filter),
	}
	return monitor.(*Metrics)
}

func (m *Metrics) RegisterFunc(id string, mFunc MonitorFunc) {

	if m.selectedMonitoringMetrics.Contains(id) && m.interval > 0*time.Second {
		m.CallerFuncs[id] = mFunc
	}
}

func (m *Metrics) RegisterCounter(id string, varType utils.CounterType) utils.Counter {

	if m.selectedMonitoringMetrics.Contains(id) && m.interval > 0*time.Second {
		if varType == utils.UintType {
			ctr := new(utils.IntCounter)
			m.Counters[id] = append(m.Counters[id], ctr)
			return ctr
		}
	}
	return &utils.NopCounter{}
}

func (m *Metrics) extractFunc(metricEvent map[string]interface{}) {
	for key, Func := range m.CallerFuncs {
		valofFunc := Func()
		splunkKey := splunkMetric + key
		metricEvent[splunkKey] = valofFunc
	}
}

func (m *Metrics) extractCounter(metricEvent map[string]interface{}) {
	for key, listofCounters := range m.Counters {
		InitSum := listofCounters[0]
		for i := 1; i < len(listofCounters); i++ {
			InitSum.Add(listofCounters[i].Value())
		}
		splunkKey := splunkMetric + key
		metricEvent[splunkKey] = InitSum.Value()
	}
}

func (m *Metrics) Start() {
	if len(m.selectedMonitoringMetrics.MapForSet) > 0 {
		ticker := time.NewTicker(m.interval)
		m.ticker = ticker

		metricEvent := make(map[string]interface{})
		for {
			select {
			case <-ticker.C:
				m.extractFunc(metricEvent)
				m.extractCounter(metricEvent)
				events := []map[string]interface{}{
					metricEvent,
				}
				m.writer.Write(events)
			}
		}
	}
}

func (m *Metrics) Stop() error {
	if m.ticker != nil {
		m.ticker.Stop()
	}
	return nil
}

func setValuesForSet(selectedMetrics string) *utils.Set {
	s := utils.NewSet()
	listofSelectedMetrics := strings.Split(selectedMetrics, ",")
	for i := 0; i < len(listofSelectedMetrics); i++ {
		s.Add(listofSelectedMetrics[i])
	}
	return s
}
