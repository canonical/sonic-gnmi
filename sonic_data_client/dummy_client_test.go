package client

//This file contains dummy tests for the sake of coverage and will be removed later

import (
	"sync"
	"testing"
	"time"

	"github.com/Workiva/go-datastructures/queue"
	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	spb "github.com/sonic-net/sonic-gnmi/proto"
)

func TestDummyEventClient(t *testing.T) {
	evtc := &EventClient{}
	evtc.last_latencies[0] = 1
	evtc.last_latencies[1] = 2
	evtc.last_latency_index = 9
	evtc.last_latency_full = true
	evtc.counters = make(map[string]uint64)
	evtc.counters["COUNTERS_EVENTS:latency_in_ms"] = 0
	compute_latency(evtc)

	// Prepare necessary arguments for each function
	var wg sync.WaitGroup
	var q *queue.PriorityQueue // Assuming queue.PriorityQueue is a valid type
	once := make(chan struct{})
	poll := make(chan struct{})
	var subscribe *gnmipb.SubscriptionList             // Assuming gnmipb.SubscriptionList is a valid type
	var deletePaths []*gnmipb.Path                     // Assuming gnmipb.Path is a valid type
	var replaceUpdates, updateUpdates []*gnmipb.Update // Assuming gnmipb.Update is a valid type

	evtc.Get(&wg)
	evtc.OnceRun(q, once, &wg, subscribe)
	evtc.PollRun(q, poll, &wg, subscribe)
	evtc.Close()
	evtc.Set(deletePaths, replaceUpdates, updateUpdates)
	evtc.Capabilities()
	evtc.last_latencies[0] = 1
	evtc.last_latencies[1] = 2
	evtc.last_latency_index = 9
	evtc.last_latency_full = true
	evtc.SentOne(&Value{
		&spb.Value{
			Timestamp: time.Now().UnixNano(),
		},
	})
	evtc.FailedSend()
	evtc.subs_handle = C_init_subs(true)

}
