// Copyright (c) 2026 Ant Group Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package diagnostics provides opt-in, process-local lifecycle timing.
package diagnostics

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Fixed buckets and call-site names bound memory. These diagnostics are not
// registered with Prometheus or exported by the ordinary metrics endpoint.
var bounds = []float64{.01, .05, .1, .25, .5, 1, 2, 5, 10, 20, 30, 40, 50, 75, 100, 150, 200, 300, 500, 750, 1000, 1500, 2000, 5000}

type timer struct {
	count   atomic.Uint64
	nanos   atomic.Uint64
	max     atomic.Uint64
	buckets [25]atomic.Uint64
}

type registry struct {
	enabled bool
	values  sync.Map
}

var timings = registry{enabled: os.Getenv("SANDBOXD_LIFECYCLE_TIMING") == "1"}

func (r *registry) observe(name string, duration time.Duration) {
	v, ok := r.values.Load(name)
	if !ok {
		v, _ = r.values.LoadOrStore(name, &timer{})
	}
	t := v.(*timer)
	ns := uint64(duration)
	t.nanos.Add(ns)
	t.count.Add(1)
	for old := t.max.Load(); ns > old && !t.max.CompareAndSwap(old, ns); old = t.max.Load() {
	}
	i := sort.SearchFloat64s(bounds, float64(duration)/float64(time.Millisecond))
	t.buckets[i].Add(1)
}

// Start accepts only fixed, code-defined phase names. Nested phases are
// inclusive; their durations must not be summed with their parent phase.
func Start(name string) func() {
	if !timings.enabled {
		return func() {}
	}
	start := time.Now()
	return func() { timings.observe(name, time.Since(start)) }
}

// Lock records acquisition wait separately from time under the lock. Record
// after releasing it so diagnostic aggregation is outside the measured lock.
func Lock(mu sync.Locker, name string) func() {
	if !timings.enabled {
		mu.Lock()
		return mu.Unlock
	}
	start := time.Now()
	mu.Lock()
	acquired := time.Now()
	return func() {
		end := time.Now()
		mu.Unlock()
		timings.observe(name+".wait", acquired.Sub(start))
		timings.observe(name+".hold", end.Sub(acquired))
	}
}

type sample struct {
	Count   uint64   `json:"count"`
	SumNS   uint64   `json:"sum_ns"`
	MaxNS   uint64   `json:"max_ns"`
	Buckets []uint64 `json:"buckets"`
}

func (r *registry) snapshot() map[string]sample {
	result := map[string]sample{}
	r.values.Range(func(key, value any) bool {
		t := value.(*timer)
		s := sample{Count: t.count.Load(), SumNS: t.nanos.Load(), MaxNS: t.max.Load(), Buckets: make([]uint64, len(bounds)+1)}
		for i := range s.Buckets {
			s.Buckets[i] = t.buckets[i].Load()
		}
		result[key.(string)] = s
		return true
	})
	return result
}

// ServeHTTP exposes cumulative counters on the existing loopback listener.
// Snapshots are individually atomic, not globally transactional. Collect idle
// snapshots around each benchmark and subtract them for a settled window.
func ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !timings.enabled {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Bounds []float64         `json:"upper_bounds_ms"`
		Stages map[string]sample `json:"stages"`
	}{bounds, timings.snapshot()})
}
