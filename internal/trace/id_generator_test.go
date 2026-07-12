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
	"go.opentelemetry.io/otel/trace"
	"reflect"
	"testing"
)

func TestGetContextID(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name  string
		args  args
		want  trace.TraceID
		want1 trace.SpanID
	}{
		{
			name: "test1",
			args: args{
				ctx: buildContext("", ""),
			},
			want:  trace.TraceID{},
			want1: trace.SpanID{},
		},
		{
			name: "test2",
			args: args{
				ctx: buildContext("f006b765f81d943aaf9339703587f409", "f2f938ec1fd47e00"),
			},
			want:  mustTraceID("f006b765f81d943aaf9339703587f409"),
			want1: mustSpanID("f2f938ec1fd47e00"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetContextID(tt.args.ctx)

			if tt.want.String() != emptyTraceID && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetContextID() got = %v, want %v", got, tt.want)
			}

			if tt.want1.String() != emptySpanID && !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetContextID() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

const (
	emptyTraceID = "00000000000000000000000000000000"
	emptySpanID  = "0000000000000000"
)

func buildContext(traceID, spanID string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextKeyTraceId, traceID)
	ctx = context.WithValue(ctx, ContextKeySpanId, spanID)
	return ctx
}

func mustTraceID(str string) trace.TraceID {
	tid, _ := trace.TraceIDFromHex(str)
	return tid
}

func mustSpanID(str string) trace.SpanID {
	sid, _ := trace.SpanIDFromHex(str)
	return sid
}
