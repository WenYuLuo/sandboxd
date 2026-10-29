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

package diagnostics

import (
	"sync"
	"testing"
	"time"
)

func TestConcurrentAccounting(t *testing.T) {
	r := registry{enabled: true}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 1000 {
				r.observe("phase", 1250*time.Microsecond)
			}
		}()
	}
	wg.Wait()
	s := r.snapshot()["phase"]
	if s.Count != 8000 || s.SumNS != 8000*1250000 || s.MaxNS != 1250000 || s.Buckets[6] != 8000 {
		t.Fatalf("lost or misbucketed concurrent samples: %+v", s)
	}
}

func TestLockPreservesMutualExclusion(t *testing.T) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	count := 0
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 1000 {
				unlock := Lock(&mu, "test.lock")
				count++
				unlock()
			}
		}()
	}
	wg.Wait()
	if count != 8000 {
		t.Fatalf("lock failed: %d", count)
	}
}
