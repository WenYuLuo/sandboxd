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

package diskusage

import "testing"

func TestUsedRatioByAvailable_Range(t *testing.T) {
	ratio, err := UsedRatioByAvailable(t.TempDir())
	if err != nil {
		t.Fatalf("UsedRatioByAvailable() error: %v", err)
	}
	if ratio < 0 || ratio > 1 {
		t.Fatalf("ratio out of range [0,1], got %f", ratio)
	}
}

func TestUsedPercentByFree_Range(t *testing.T) {
	percent, err := UsedPercentByFree(t.TempDir())
	if err != nil {
		t.Fatalf("UsedPercentByFree() error: %v", err)
	}
	if percent < 0 || percent > 100 {
		t.Fatalf("percent out of range [0,100], got %f", percent)
	}
}
