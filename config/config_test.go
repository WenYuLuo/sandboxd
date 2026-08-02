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

package config

import "testing"

func TestNormalizeCPULimitMode(t *testing.T) {
	for _, test := range []struct {
		input   string
		want    string
		wantErr bool
	}{
		{input: "", want: CPULimitModeShares},
		{input: " shares ", want: CPULimitModeShares},
		{input: "QUOTA", want: CPULimitModeQuota},
		{input: "cpuset", wantErr: true},
	} {
		got, err := NormalizeCPULimitMode(test.input)
		if test.wantErr {
			if err == nil {
				t.Errorf("NormalizeCPULimitMode(%q) unexpectedly succeeded", test.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("NormalizeCPULimitMode(%q): %v", test.input, err)
			continue
		}
		if got != test.want {
			t.Errorf("NormalizeCPULimitMode(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}
