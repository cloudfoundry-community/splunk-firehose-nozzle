package eventsource

import (
	"code.cloudfoundry.org/go-loggregator/v8"
	"code.cloudfoundry.org/go-loggregator/v8/conversion"
	"code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"
	"code.cloudfoundry.org/lager"
	"context"
	"github.com/cloudfoundry/sonde-go/events"
	"sync/atomic"
	"time"
)

// Streamer implements Stream which returns a new EnvelopeStream for the given context and request.
type Streamer interface {
	// EnvelopeStream returns batches of envelopes.
	Stream(ctx context.Context, req *loggregator_v2.EgressBatchRequest) loggregator.EnvelopeStream
}

// V2Adapter struct with field of type streamer
type V2Adapter struct {
	streamer Streamer
}

// NewV2Adapter returns v2Adapter
func NewV2Adapter(s Streamer) V2Adapter {
	return V2Adapter{
		streamer: s,
	}
}

// Firehose returns only selected event stream
func (a V2Adapter) Firehose(config *FirehoseConfig) chan *events.Envelope {
	ctx := context.Background()
	var v1msgs = make(chan *events.Envelope, 10000)
	var v2msgs = make(chan []*loggregator_v2.Envelope, 10000)
	es := a.streamer.Stream(ctx, &loggregator_v2.EgressBatchRequest{
		ShardId: config.SubscriptionID,
		Selectors: []*loggregator_v2.Selector{
			{
				Message: &loggregator_v2.Selector_Log{
					Log: &loggregator_v2.LogSelector{},
				},
			},
			{
				Message: &loggregator_v2.Selector_Counter{
					Counter: &loggregator_v2.CounterSelector{},
				},
			},
			{
				Message: &loggregator_v2.Selector_Event{
					Event: &loggregator_v2.EventSelector{},
				},
			},
			{
				Message: &loggregator_v2.Selector_Gauge{
					Gauge: &loggregator_v2.GaugeSelector{},
				},
			},
			{
				Message: &loggregator_v2.Selector_Timer{
					Timer: &loggregator_v2.TimerSelector{},
				},
			},
		},
	})

	go func() {
		for ctx.Err() == nil {
			e := es()
			v2msgs <- e
		}
	}()

	go func() {
		var receivedCount uint64 = 0

		if config.StatusMonitorInterval > time.Second*0 {
			timer := time.NewTimer(config.StatusMonitorInterval)
			for ctx.Err() == nil {
				select {
				case <-timer.C:
					config.Logger.Info("Event_Count", lager.Data{"event_count_received": receivedCount})
					timer.Reset(config.StatusMonitorInterval)
					receivedCount = 0
				default:
				}
				select {
				case eArray := <-v2msgs:
					atomic.AddUint64(&receivedCount, uint64(len(eArray)))
					for _, e := range eArray {
						//// ToV1 converts v2 envelopes down to v1 envelopes.
						for _, v1e := range conversion.ToV1(e) {
							v1msgs <- v1e
						}
					}
				default:
				}
			}
		} else {
			for ctx.Err() == nil {
				eArray := <-v2msgs
				for _, e := range eArray {
					for _, v1e := range conversion.ToV1(e) {
						v1msgs <- v1e
					}
				}
			}
		}
	}()

	return v1msgs
}
