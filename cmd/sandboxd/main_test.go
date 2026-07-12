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

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsMuxDoesNotExposePprof(t *testing.T) {
	metrics := httptest.NewRecorder()
	metricsMux().ServeHTTP(metrics, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if metrics.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want %d", metrics.Code, http.StatusOK)
	}

	pprof := httptest.NewRecorder()
	metricsMux().ServeHTTP(pprof, httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil))
	if pprof.Code != http.StatusNotFound {
		t.Fatalf("metrics mux exposed pprof with status %d", pprof.Code)
	}
}

func TestPprofMuxDoesNotExposeMetrics(t *testing.T) {
	pprof := httptest.NewRecorder()
	pprofMux().ServeHTTP(pprof, httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil))
	if pprof.Code != http.StatusOK {
		t.Fatalf("GET /debug/pprof/ status = %d, want %d", pprof.Code, http.StatusOK)
	}

	metrics := httptest.NewRecorder()
	pprofMux().ServeHTTP(metrics, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if metrics.Code != http.StatusNotFound {
		t.Fatalf("pprof mux exposed metrics with status %d", metrics.Code)
	}
}
