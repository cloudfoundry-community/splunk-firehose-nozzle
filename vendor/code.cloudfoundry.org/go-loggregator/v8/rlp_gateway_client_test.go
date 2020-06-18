package loggregator_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang/protobuf/jsonpb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"golang.org/x/net/context"

	"code.cloudfoundry.org/go-loggregator/v8"
	"code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"
)

var _ = Describe("RlpGatewayClient", func() {
	var (
		spyDoer   *spyDoer
		c         *loggregator.RLPGatewayClient
		logBuffer *gbytes.Buffer
	)

	BeforeEach(func() {
		spyDoer = newSpyDoer()
		logBuffer = gbytes.NewBuffer()
		c = loggregator.NewRLPGatewayClient(
			"https://some.addr",
			loggregator.WithRLPGatewayHTTPClient(spyDoer),
			loggregator.WithRLPGatewayClientLogger(log.New(logBuffer, "", 0)),
		)
	})

	It("requests envelopes from the RLP", func() {
		ch := make(chan []byte, 100)
		spyDoer.resps = append(spyDoer.resps, &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(channelReader(ch)),
		})
		spyDoer.errs = []error{nil}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c.Stream(ctx, &loggregator_v2.EgressBatchRequest{
			ShardId:           "some-shard",
			DeterministicName: "some-name",
			Selectors: []*loggregator_v2.Selector{
				{SourceId: "some-id-1", Message: &loggregator_v2.Selector_Log{Log: &loggregator_v2.LogSelector{}}},
			},
		})

		Eventually(spyDoer.Reqs).Should(HaveLen(1))

		req := spyDoer.Reqs()[0]

		Expect(req.URL.Scheme).To(Equal("https"))
		Expect(req.URL.Host).To(Equal("some.addr"))
		Expect(req.URL.Path).To(Equal("/v2/read"))
		Expect(req.Header.Get("Accept")).To(Equal("text/event-stream"))
		Expect(req.Header.Get("Cache-Control")).To(Equal("no-cache"))
		Expect(req.Method).To(Equal(http.MethodGet))
		Expect(req.URL.Query().Get("shard_id")).To(Equal("some-shard"))
		Expect(req.URL.Query().Get("deterministic_name")).To(Equal("some-name"))
		Expect(req.URL.Query().Get("source_id")).To(Equal("some-id-1"))
		Expect(req.URL.Query()).To(HaveKey("log"))
	})

	DescribeTable("encodes selectors correctly",
		func(selectors []*loggregator_v2.Selector, paramKey string, paramValue []string) {
			ch := make(chan []byte, 100)
			spyDoer.resps = append(spyDoer.resps, &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(channelReader(ch)),
			})
			spyDoer.errs = []error{nil}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			c.Stream(ctx, &loggregator_v2.EgressBatchRequest{
				ShardId:           "some-shard",
				DeterministicName: "some-name",
				Selectors:         selectors,
			})

			Eventually(spyDoer.Reqs).Should(HaveLen(1))

			req := spyDoer.Reqs()[0]
			Expect(req.URL.Query()).To(HaveKeyWithValue(paramKey, paramValue))
			Expect(req.URL.Query()).To(HaveLen(3))
		},
		Entry("log",
			[]*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Log{
						Log: &loggregator_v2.LogSelector{},
					},
				},
			}, "log", []string{""}),
		Entry("counter",
			[]*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Counter{
						Counter: &loggregator_v2.CounterSelector{},
					},
				},
			}, "counter", []string{""}),
		Entry("counter and counter with name",
			[]*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Counter{
						Counter: &loggregator_v2.CounterSelector{},
					},
				},
				{
					Message: &loggregator_v2.Selector_Counter{
						Counter: &loggregator_v2.CounterSelector{
							Name: "cntr",
						},
					},
				},
			}, "counter.name", []string{"cntr"}),
		Entry("gauge", []*loggregator_v2.Selector{
			{
				Message: &loggregator_v2.Selector_Gauge{
					Gauge: &loggregator_v2.GaugeSelector{},
				},
			},
		}, "gauge", []string{""}),
		Entry("gauge with name",
			[]*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Gauge{
						Gauge: &loggregator_v2.GaugeSelector{}},
				},
				{
					Message: &loggregator_v2.Selector_Gauge{
						Gauge: &loggregator_v2.GaugeSelector{
							Names: []string{"gauge"},
						},
					},
				},
			}, "gauge.name", []string{"gauge"}),
		Entry("timer",
			[]*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Timer{
						Timer: &loggregator_v2.TimerSelector{},
					},
				},
			}, "timer", []string{""}),
		Entry("event",
			[]*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Event{
						Event: &loggregator_v2.EventSelector{},
					},
				},
			}, "event", []string{""}),
		Entry("many source ID",
			[]*loggregator_v2.Selector{
				{
					SourceId: "some-id-1",
				},
				{
					SourceId: "some-id-2",
				},
				{
					SourceId: "some-id-2",
				},
			}, "source_id", []string{"some-id-1", "some-id-2"}),
	)

	It("streams envelopes", func() {
		ch := make(chan []byte, 100)
		spyDoer.resps = append(spyDoer.resps, &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(channelReader(ch)),
		})
		spyDoer.errs = []error{nil}

		go func() {
			m := jsonpb.Marshaler{}
			for i := 0; i < 10; i++ {
				s, err := m.MarshalToString(&loggregator_v2.EnvelopeBatch{
					Batch: []*loggregator_v2.Envelope{
						{Timestamp: int64(i)},
					},
				})
				if err != nil {
					panic(err)
				}
				ch <- []byte(fmt.Sprintf("data: %s\n\n", s))
			}
		}()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		es := c.Stream(ctx, &loggregator_v2.EgressBatchRequest{})

		envelopes := make(chan *loggregator_v2.Envelope, 100)
		go func() {
			for ctx.Err() == nil {
				for _, e := range es() {
					envelopes <- e
				}
			}
		}()

		Eventually(envelopes).Should(HaveLen(10))
	})

	It("handles heartbeats", func() {
		ch := make(chan []byte, 100)
		spyDoer.resps = append(spyDoer.resps, &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(channelReader(ch)),
		})
		spyDoer.errs = []error{nil}

		go func() {
			for i := 0; i < 10; i++ {
				ch <- []byte("event: heartbeat\ndata: 1541438163\n\n")
			}
		}()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c.Stream(ctx, &loggregator_v2.EgressBatchRequest{})

		// TODO: Asserting on the logs is far from ideal, however the only
		// output from an unmarshalling error is a log line. If we decide to
		// do more with an error (e.g., metrics), this test should be
		// adjusted.
		Consistently(logBuffer.Contents).Should(BeEmpty())
	})

	It("handles closing events", func() {
		ch := make(chan []byte, 100)
		noCloseCh := make(chan []byte, 100)
		spyDoer.resps = append(spyDoer.resps,
			&http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(channelReader(ch)),
			},
			&http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(channelReader(noCloseCh)),
			})
		spyDoer.errs = []error{nil, nil}

		ch <- []byte("event: closing\ndata: message\n\n")
		close(ch)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		c.Stream(ctx, &loggregator_v2.EgressBatchRequest{})

		Eventually(func() int {
			return len(spyDoer.Reqs())
		}).Should(BeNumerically("==", 2))
	})

	It("reconnects for non-200 requests", func() {
		spyDoer.resps = append(spyDoer.resps, &http.Response{StatusCode: 500})
		spyDoer.resps = append(spyDoer.resps, &http.Response{StatusCode: 500})
		spyDoer.resps = append(spyDoer.resps, &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(channelReader(nil)),
		})
		spyDoer.errs = []error{nil, nil, nil}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		c.Stream(ctx, &loggregator_v2.EgressBatchRequest{})

		Eventually(spyDoer.Reqs).Should(HaveLen(3))
	})

	It("reconnects for any errors", func() {
		spyDoer.resps = append(spyDoer.resps, &http.Response{StatusCode: 200})
		spyDoer.errs = append(spyDoer.errs, errors.New("some-error"))

		spyDoer.resps = append(spyDoer.resps, &http.Response{StatusCode: 200})
		spyDoer.errs = append(spyDoer.errs, errors.New("some-error"))

		spyDoer.resps = append(spyDoer.resps, &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(channelReader(nil)),
		})
		spyDoer.errs = append(spyDoer.errs, nil)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		c.Stream(ctx, &loggregator_v2.EgressBatchRequest{})

		Eventually(spyDoer.Reqs).Should(HaveLen(3))
		Consistently(spyDoer.Reqs).Should(HaveLen(3))
	})

	It("stops trying to reconnect after maxRetries", func() {
		c = loggregator.NewRLPGatewayClient(
			"https://some.addr",
			loggregator.WithRLPGatewayHTTPClient(spyDoer),
			loggregator.WithRLPGatewayClientLogger(log.New(logBuffer, "", 0)),
			loggregator.WithRLPGatewayMaxRetries(3),
		)
		spyDoer.onlyErrs = true

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		es := c.Stream(ctx, &loggregator_v2.EgressBatchRequest{})

		Eventually(es).Should(BeNil())

		Expect(len(spyDoer.Reqs())).To(Equal(4)) //1 initial try + 3 retries
	})

	It("puts error on error channel after maxRetries", func() {
		errChan := make(chan error, 10)

		c = loggregator.NewRLPGatewayClient(
			"https://some.addr",
			loggregator.WithRLPGatewayHTTPClient(spyDoer),
			loggregator.WithRLPGatewayClientLogger(log.New(logBuffer, "", 0)),
			loggregator.WithRLPGatewayMaxRetries(3),
			loggregator.WithRLPGatewayErrChan(errChan),
		)
		spyDoer.onlyErrs = true

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		es := c.Stream(ctx, &loggregator_v2.EgressBatchRequest{})

		Eventually(es).Should(BeNil())

		var err error
		Eventually(errChan).Should(Receive(&err))
		Expect(err).To(MatchError("client connection attempts exceeded max retries -- giving up"))
	})

	It("doesn't fail when the passed error chan is full", func() {
		errChan := make(chan error, 1)
		errChan <- errors.New("something that happened before")

		c = loggregator.NewRLPGatewayClient(
			"https://some.addr",
			loggregator.WithRLPGatewayHTTPClient(spyDoer),
			loggregator.WithRLPGatewayClientLogger(log.New(logBuffer, "", 0)),
			loggregator.WithRLPGatewayMaxRetries(3),
			loggregator.WithRLPGatewayErrChan(errChan),
		)
		spyDoer.onlyErrs = true

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		es := c.Stream(ctx, &loggregator_v2.EgressBatchRequest{})
		esClosed := make(chan struct{})
		go func() {
			es()
			esClosed <- struct{}{}
		}()

		Eventually(esClosed).Should(Receive())
	})

	It("resets connection attempts when connection succeeds", func() {
		c = loggregator.NewRLPGatewayClient(
			"https://some.addr",
			loggregator.WithRLPGatewayHTTPClient(spyDoer),
			loggregator.WithRLPGatewayClientLogger(log.New(logBuffer, "", 0)),
			loggregator.WithRLPGatewayMaxRetries(2),
		)

		errorResponse(spyDoer)
		errorResponse(spyDoer)

		ch := make(chan []byte, 100)
		spyDoer.resps = append(spyDoer.resps, &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(channelReader(ch)),
		})
		spyDoer.errs = append(spyDoer.errs, nil)

		ch <- []byte("event: closing\ndata: message\n\n")
		close(ch)

		errorResponse(spyDoer)
		errorResponse(spyDoer)
		errorResponse(spyDoer)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		es := c.Stream(ctx, &loggregator_v2.EgressBatchRequest{})
		Eventually(es).Should(BeNil())

		Expect(len(spyDoer.Reqs())).To(Equal(6))
	})
})

func errorResponse(sd *spyDoer) {
	sd.resps = append(sd.resps, &http.Response{StatusCode: 200})
	sd.errs = append(sd.errs, errors.New("some-error"))
}

type spyDoer struct {
	mu       sync.Mutex
	reqs     []*http.Request
	resps    []*http.Response
	errs     []error
	onlyErrs bool
}

func newSpyDoer() *spyDoer {
	return &spyDoer{}
}

func (s *spyDoer) Do(r *http.Request) (*http.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reqs = append(s.reqs, r)

	if len(s.resps) != len(s.errs) {
		panic("out of sync")
	}

	if s.onlyErrs {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(nil)),
		}, errors.New("default error")
	}

	if len(s.resps) == 0 {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(nil)),
		}, nil
	}

	resp, err := s.resps[0], s.errs[0]
	s.resps, s.errs = s.resps[1:], s.errs[1:]

	if resp.Body == nil {
		resp.Body = ioutil.NopCloser(bytes.NewReader(nil))
	}

	return resp, err
}

func (s *spyDoer) Reqs() []*http.Request {
	s.mu.Lock()
	defer s.mu.Unlock()

	results := make([]*http.Request, len(s.reqs))
	copy(results, s.reqs)
	return results
}

type channelReader <-chan []byte

func (r channelReader) Read(buf []byte) (int, error) {
	data, ok := <-r
	if !ok {
		return 0, io.EOF
	}
	n := copy(buf, data)
	return n, nil
}
