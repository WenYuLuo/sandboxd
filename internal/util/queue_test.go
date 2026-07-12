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

package util

import (
	"fmt"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	queue1 := New[int](0)
	queue2 := New[string]("empty")

	if queue1.Pop() != 0 {
		t.Errorf("queue1 pop error")
	}

	if queue2.Pop() != "empty" {
		t.Errorf("queue2 pop error")
	}

	queue1.Push(1)
	queue2.Push("one")

	if queue1.Get(0) != 1 {
		t.Errorf("queue1 get error")
	}

	if queue2.Get(0) != "one" {
		t.Errorf("queue2 get error")
	}

	if queue1.Pop() != 1 {
		t.Errorf("queue1 pop error")
	}

	if queue2.Pop() != "one" {
		t.Errorf("queue2 pop error")
	}

	if queue1.Pop() != 0 {
		t.Errorf("queue1 pop error")
	}

	if queue2.Pop() != "empty" {
		t.Errorf("queue2 pop error")
	}
}

// write a benchmark function for queue
func BenchmarkQueueOp(b *testing.B) {
	operationQps := []int{
		100,
		500,
		10000,
		300000,
	}
	b.ResetTimer()
	for _, qps := range operationQps {
		b.StopTimer()
		b.Run(fmt.Sprintf("qps: %v", qps), func(b *testing.B) {
			// prepare data
			queue1 := New[int](0)
			for i := 0; i < 1000; i++ {
				queue1.Push(i)
			}
			// check data.
			if queue1.Length() != 1000 {
				b.Errorf("queue1 length error")
			}

			popResult := make([]int, 0)

			b.StartTimer()
			var opsWg sync.WaitGroup
			opsWg.Add(4)
			// pop
			go func(queue *Queue[int], res []int) {
				defer opsWg.Done()

				var internalWg sync.WaitGroup
				for i := 0; i < qps; i++ {
					internalWg.Add(1)
					go func(queue *Queue[int]) {
						defer internalWg.Done()
						res = append(res, queue.Pop())
					}(queue)
				}
				internalWg.Wait()
			}(queue1, popResult)

			b.StopTimer()
			// check result
			resMap := make(map[int]int)
			for _, v := range popResult {
				resMap[v]++
			}
			for _, v := range resMap {
				if v != 1 {
					b.Errorf("queue1 result error")
				}
			}
			b.StartTimer()

			// top
			go func(queue *Queue[int]) {
				defer opsWg.Done()

				var internalWg sync.WaitGroup
				for i := 0; i < qps; i++ {
					internalWg.Add(1)
					go func(queue *Queue[int]) {
						defer internalWg.Done()
						queue.Top()
					}(queue)
				}
				internalWg.Wait()
			}(queue1)

			// read
			go func(queue *Queue[int]) {
				defer opsWg.Done()

				var internalWg sync.WaitGroup
				for i := 0; i < qps; i++ {
					internalWg.Add(1)
					go func(queue *Queue[int]) {
						defer internalWg.Done()
						queue.Length()
					}(queue)
				}
				internalWg.Wait()
			}(queue1)

			// push
			go func(queue *Queue[int]) {
				defer opsWg.Done()

				var internalWg sync.WaitGroup
				for i := 0; i < qps; i++ {
					internalWg.Add(1)
					go func(queue *Queue[int], num int) {
						defer internalWg.Done()
						queue.Push(num)
					}(queue, i)
				}
				internalWg.Wait()
			}(queue1)
			opsWg.Wait()
		})
	}
}
