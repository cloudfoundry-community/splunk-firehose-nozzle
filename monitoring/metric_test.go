package monitoring_test

import (
	"time"

	"code.cloudfoundry.org/lager"
	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/monitoring"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/testing"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Monitoring", func() {
	var (
		selectedMonitoringMetrics string = "nozzle.queue.percentage,splunk.events.dropped.count,splunk.events.sent.count"
		writer                    testing.EventWriterMetricMock
		Counter                   utils.Counter
		monitor                   = NewMetricsMonitor(lager.NewLogger("Test"), 2*time.Second, &writer, selectedMonitoringMetrics)
	)

	BeforeEach(func() {
		monitor.RegisterFunc("nozzle.queue.percentage", func() interface{} { return 10 })
		Counter = monitor.RegisterCounter("splunk.events.sent.count", utils.UintType)

	})

	It("Test Monitor Func", func() {
		checkLen := len(monitor.(*Metrics).CallerFuncs)
		Expect(checkLen).To(Equal(1))

	})

	It("Test Register Counter", func() {
		checkLen := len(monitor.(*Metrics).Counters)
		Expect(checkLen).To(Equal(1))

	})

	It("Test of Run", func() {
		Counter.Add(10)
		go monitor.Start()
		time.Sleep(3 * time.Second)
		monitor.Stop()
		Expect(writer.CapturedEvents[len(writer.CapturedEvents)-1]["metric_name:splunk.events.sent.count"]).To(Equal(uint64(10)))
		Expect(len(writer.CapturedEvents)).To(Equal(1))
	})

	It("Identical Key ", func() {
		Counter.Add(10)
		Counter := RegisterCounter("splunk.events.sent.count", utils.UintType)
		Counter.Add(10)
		go monitor.Start()
		time.Sleep(3 * time.Second)
		monitor.Stop()
		Expect(writer.CapturedEvents[len(writer.CapturedEvents)-1]["metric_name:splunk.events.sent.count"]).To(Equal(uint64(40)))
	})

	It("Test when metric is disabled", func() {
		monitor = NewMetricsMonitor(lager.NewLogger("Test"), 0*time.Second, &writer, selectedMonitoringMetrics)
		checkLenCounter := len(monitor.(*Metrics).Counters)
		checkLenFuncs := len(monitor.(*Metrics).CallerFuncs)
		Expect(checkLenFuncs).To(Equal(0))
		Expect(checkLenCounter).To(Equal(0))

	})

})
