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

import "sync"

// minQueueLen is smallest capacity that queue may have.
const minQueueLen = 64

// Queue represents a single instance of the queue data structure.
type Queue[T comparable] struct {
	buf []*T
	// ensures elements uniq
	buf_map map[T]struct{}
	// empty is the value that is returned when the queue is empty.
	empty             T
	head, tail, count int
	sync.Mutex
}

// New constructs and returns a new Queue.
func New[T comparable](emptyObj T) *Queue[T] {
	return &Queue[T]{
		buf:     make([]*T, minQueueLen),
		buf_map: make(map[T]struct{}, minQueueLen),
		empty:   emptyObj,
	}
}

// Length returns the number of elements currently stored in the queue.
func (q *Queue[T]) Length() int {
	q.Lock()
	defer q.Unlock()
	return q.count
}

// resizes the queue to fit exactly twice its current contents
// this can result in shrinking if the queue is less than half-full
func (q *Queue[T]) resize() {
	newBuf := make([]*T, q.count<<1)

	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}

	q.head = 0
	q.tail = q.count
	q.buf = newBuf
}

func (q *Queue[T]) push_unlocked(elem T) {
	if q.count == len(q.buf) {
		q.resize()
	}

	// key action to ensure elements in queue are uniq
	q.buf_map[elem] = struct{}{}

	q.buf[q.tail] = &elem
	// bitwise modulus
	q.tail = (q.tail + 1) & (len(q.buf) - 1)
	q.count++
}

// Push puts an element on the end of the queue.
func (q *Queue[T]) Push(elem T) {
	q.Lock()
	defer q.Unlock()

	// Only do the push if there is no same one in the queue.
	if !q.has_unlocked(elem) {
		q.push_unlocked(elem)
	}
}

func (q *Queue[T]) has_unlocked(elem T) bool {
	_, has := q.buf_map[elem]
	return has
}

// Has returns whether the input element exists
func (q *Queue[T]) Has(elem T) bool {
	q.Lock()
	defer q.Unlock()

	return q.has_unlocked(elem)
}

// Top returns the element at the head of the queue.
func (q *Queue[T]) Top() T {
	q.Lock()
	defer q.Unlock()
	if q.count <= 0 {
		return q.empty
	}
	return *(q.buf[q.head])
}

// Get returns the element at index i in the queue. If the index is
// invalid, the call will panic. This method accepts both positive and
// negative index values. Index 0 refers to the first element, and
// index -1 refers to the last.
func (q *Queue[T]) Get(i int) T {
	q.Lock()
	defer q.Unlock()
	// If indexing backwards, convert to positive index.
	if i < 0 {
		i += q.count
	}
	if i < 0 || i >= q.count {
		return q.empty
	}
	// bitwise modulus
	return *(q.buf[(q.head+i)&(len(q.buf)-1)])
}

// Pop removes and returns the element from the front of the queue.
// Returns the empty value if the queue is empty.
func (q *Queue[T]) Pop() T {
	q.Lock()
	defer q.Unlock()
	if q.count <= 0 {
		return q.empty
	}
	ret := q.buf[q.head]
	q.buf[q.head] = nil
	// bitwise modulus
	q.head = (q.head + 1) & (len(q.buf) - 1)
	q.count--
	// Resize down if buffer 1/4 full.
	if len(q.buf) > minQueueLen && (q.count<<2) == len(q.buf) {
		q.resize()
	}

	delete(q.buf_map, *ret)

	return *ret
}

func (q *Queue[T]) List() []T {
	q.Lock()
	defer q.Unlock()
	list := make([]T, 0, q.count)
	for i := 0; i < q.count; i++ {
		// bitwise modulus, mirrors Get without re-locking
		list = append(list, *(q.buf[(q.head+i)&(len(q.buf)-1)]))
	}
	return list
}

func (q *Queue[T]) Num() int {
	q.Lock()
	defer q.Unlock()
	return q.count
}
