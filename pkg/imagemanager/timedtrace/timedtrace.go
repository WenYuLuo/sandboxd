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

package timedtrace

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Config describes how a timed operation should be traced and logged.
type Config struct {
	TracerName      string
	Operation       string
	IdentifierKey   string
	IdentifierValue string
	LogPrefix       string
	Attributes      []attribute.KeyValue
}

// Operation wraps a traced operation and records stage timings.
type Operation struct {
	span            trace.Span
	operation       string
	identifierKey   string
	identifierValue string
	logPrefix       string
	startTime       time.Time
	stages          []stageRecord
}

type stageRecord struct {
	name     string
	duration time.Duration
}

// Start begins a timed operation.
func Start(ctx context.Context, cfg Config) (*Operation, context.Context) {
	tracerName := cfg.TracerName
	if tracerName == "" {
		tracerName = "timedtrace"
	}
	tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer(tracerName)
	ctx, span := tracer.Start(ctx, cfg.Operation, trace.WithAttributes(cfg.Attributes...))

	return &Operation{
		span:            span,
		operation:       cfg.Operation,
		identifierKey:   cfg.IdentifierKey,
		identifierValue: cfg.IdentifierValue,
		logPrefix:       cfg.LogPrefix,
		startTime:       time.Now(),
		stages:          make([]stageRecord, 0, 8),
	}, ctx
}

// Stage records a stage timing and emits a span event.
func (o *Operation) Stage(stageName string, duration time.Duration) {
	o.stages = append(o.stages, stageRecord{name: stageName, duration: duration})
	if o.span.IsRecording() {
		o.span.AddEvent(stageName,
			trace.WithAttributes(
				attribute.String("stage", stageName),
				attribute.Int64("duration_ms", duration.Milliseconds()),
			),
		)
	}
}

// RecordError records an error on the span.
func (o *Operation) RecordError(err error) {
	if err != nil {
		o.span.RecordError(err)
	}
}

// End closes the span and logs timing summary.
func (o *Operation) End() {
	defer o.span.End()

	totalDuration := time.Since(o.startTime)
	safeTotal := totalDuration
	if safeTotal <= 0 {
		safeTotal = time.Nanosecond
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("operation=%s", o.operation))
	if o.identifierKey != "" {
		summary.WriteString(fmt.Sprintf(" %s=%s", o.identifierKey, o.identifierValue))
	}
	summary.WriteString(fmt.Sprintf(" total=%v", totalDuration))
	for _, stage := range o.stages {
		percentage := float64(stage.duration) / float64(safeTotal) * 100
		summary.WriteString(fmt.Sprintf(", %s=%v(%.1f%%)", stage.name, stage.duration, percentage))
	}

	if o.logPrefix != "" {
		logrus.Infof("%s %s", o.logPrefix, summary.String())
	} else {
		logrus.Infof("%s", summary.String())
	}
	o.span.SetAttributes(attribute.Int64("total_duration_ms", totalDuration.Milliseconds()))
}

// Fail records an error and ends the operation.
func (o *Operation) Fail(err error) {
	o.RecordError(err)
	o.End()
}
