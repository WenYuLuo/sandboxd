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

package trace

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sync"

	. "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	ContextKeyTraceId = "TRACEID"
	ContextKeySpanId  = "SPANID"
)

var defaultIDGenerator IDGenerator

func init() {
	defaultIDGenerator = newRandomTraceIDGenerator()
}

type randomTraceIDGenerator struct {
	sync.Mutex
	randSource *rand.Rand
}

var _ IDGenerator = &randomTraceIDGenerator{}

// NewSpanID returns a non-zero span ID from a randomly-chosen sequence.
func (gen *randomTraceIDGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	gen.Lock()
	defer gen.Unlock()
	sid := trace.SpanID{}
	_, _ = gen.randSource.Read(sid[:])
	return sid
}

// NewIDs returns a non-zero trace ID and a non-zero span ID from a
// randomly-chosen sequence.
func (gen *randomTraceIDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	gen.Lock()
	defer gen.Unlock()
	tid := trace.TraceID{}
	_, _ = gen.randSource.Read(tid[:])
	sid := trace.SpanID{}
	_, _ = gen.randSource.Read(sid[:])
	return tid, sid
}

func newRandomTraceIDGenerator() IDGenerator {
	gen := &randomTraceIDGenerator{
		Mutex: sync.Mutex{},
	}
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	gen.randSource = rand.New(rand.NewSource(rngSeed))
	return gen
}

func GetContextID(ctx context.Context) (trace.TraceID, trace.SpanID) {
	return GetTraceIdFromContext(ctx), GetSpanIdFromContext(ctx)
}

// GetTraceIdFromContext try to get trace id from context, if not exist, generate a new one
func GetTraceIdFromContext(ctx context.Context) trace.TraceID {
	if id, ok := ctx.Value(ContextKeyTraceId).(string); ok {
		if tid, err := trace.TraceIDFromHex(id); err == nil {
			return tid
		}
	}
	tid, _ := defaultIDGenerator.NewIDs(context.Background())
	return tid
}

// GetSpanIdFromContext try to get span id from context, if not exist, generate a new one
func GetSpanIdFromContext(ctx context.Context) trace.SpanID {
	if id, ok := ctx.Value(ContextKeySpanId).(string); ok {
		if sid, err := trace.SpanIDFromHex(id); err == nil {
			return sid
		}
	}
	_, sid := defaultIDGenerator.NewIDs(context.Background())
	return sid
}
