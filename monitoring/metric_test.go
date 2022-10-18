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
		filtermetric string = "nozzle.queue.percentage,splunk.events.dropped.count,splunk.events.sent.count,firehose.events.dropped.count,firehose.events.received.count,splunk.events.throughput,nozzle.usage.ram,nozzle.usage.cpu,firehose.events.received.count.normalreceive,nozzle.cachehit.memory,nozzle.cachemiss.memory,nozzle.cachehit.remote,nozzle.cachemiss.remote,nozzle.cachehit.boltdb,nozzle.cachemiss.boltdb"
		writer       testing.EventWriterMetricMock
		monitor      = InitMetrics(lager.NewLogger("Test"), 2*time.Second, &writer, filtermetric)
		Counter      utils.Counter
	)

	BeforeEach(func() {
		RegisterFunc("nozzle.queue.percentage", func() interface{} { return 10 })
		Counter = RegisterCounter("splunk.events.sent.count", utils.UintType)
		Counter.Add(10)
	})

	It("Test Monitor Func", func() {
		checkLen := len(monitor.CallerFuncs)
		Expect(checkLen).To(Equal(1))

	})

	It("Test Register Counter", func() {
		checkLen := len(monitor.Counters)
		Expect(checkLen).To(Equal(1))

	})

	It("Test of Run", func() {
		go monitor.Start()
		time.Sleep(3 * time.Second)
		monitor.Stop()
		Expect(writer.CapturedEvents[len(writer.CapturedEvents)-1]["metric_name:splunk.events.sent.count"]).To(Equal(uint64(30)))
		Expect(len(writer.CapturedEvents)).To(Equal(1))
	})

	It("Identical Key ", func() {
		Counter := RegisterCounter("splunk.events.sent.count", utils.UintType)
		Counter.Add(10)
		go monitor.Start()
		time.Sleep(3 * time.Second)
		monitor.Stop()
		Expect(writer.CapturedEvents[len(writer.CapturedEvents)-1]["metric_name:splunk.events.sent.count"]).To(Equal(uint64(20)))
	})

	It("Test when metric is disabled", func() {
		monitor = InitMetrics(lager.NewLogger("Test"), 0*time.Second, &writer, filtermetric)
		checkLenCounter := len(monitor.Counters)
		checkLenFuncs := len(monitor.CallerFuncs)
		Expect(checkLenFuncs).To(Equal(0))
		Expect(checkLenCounter).To(Equal(0))

	})

})
