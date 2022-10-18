package monitoring

import "github.com/cloudfoundry-community/splunk-firehose-nozzle/utils"

type NoMonitor struct {
}

func NewNoMonitor() Monitor {
	return &NoMonitor{}
}

func (nm *NoMonitor) RegisterFunc(id string, caller MonitorFunc) {
	return
}

func (nm *NoMonitor) RegisterCounter(id string, varType utils.CounterType) utils.Counter {
	return &utils.NopCounter{}
}
func (nm *NoMonitor) Start() error {
	return nil
}

func (nm *NoMonitor) Stop() error {
	return nil
}
