package monitoring

import "github.com/cloudfoundry-community/splunk-firehose-nozzle/utils"

type NoMonitor struct {
}

func NewNoMonitor() Monitor {
	if monitor != nil {
		monitor.Stop()
	}
	return &NoMonitor{}
}

func (nm *NoMonitor) RegisterFunc(id string, caller MonitorFunc) {
	return
}

func (nm *NoMonitor) RegisterCounter(id string, varType utils.CounterType) utils.Counter {
	return &utils.NopCounter{}
}
func (nm *NoMonitor) Start() {
}

func (nm *NoMonitor) Stop() error {
	return nil
}
