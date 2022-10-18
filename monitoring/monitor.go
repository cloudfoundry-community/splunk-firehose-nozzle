package monitoring

import "github.com/cloudfoundry-community/splunk-firehose-nozzle/utils"

type MonitorFunc func() interface{}

type Monitor interface {
	RegisterFunc(string, MonitorFunc)
	RegisterCounter(string, utils.CounterType) utils.Counter
	Start() error
	Stop() error
}
