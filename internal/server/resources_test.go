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

package server

import (
	"testing"

	"github.com/inclusionAI/sandboxd/config"
)

func TestResourcesToLinuxCPULimitModes(t *testing.T) {
	shares := resourcesToLinux(map[string]float64{
		"CPU":    1000,
		"Memory": 2048,
	}, config.CPULimitModeShares)
	if shares.CpuShares != 1024 || shares.CpuQuota != 0 || shares.CpuPeriod != 0 {
		t.Fatalf("shares resources = %+v", shares)
	}

	quota := resourcesToLinux(map[string]float64{
		"CPU":    1000,
		"Memory": 2048,
	}, config.CPULimitModeQuota)
	if quota.CpuShares != 0 || quota.CpuQuota != 100000 || quota.CpuPeriod != 100000 {
		t.Fatalf("quota resources = %+v", quota)
	}
	if quota.MemoryLimitInBytes != 2*1024*1024*1024 {
		t.Fatalf("quota memory limit = %d", quota.MemoryLimitInBytes)
	}
}

func TestResourcesToLinuxQuotaDefaultsAndFractionalCPU(t *testing.T) {
	defaults := resourcesToLinux(nil, config.CPULimitModeQuota)
	if defaults.CpuShares != 0 || defaults.CpuQuota != 50000 || defaults.CpuPeriod != 100000 {
		t.Fatalf("default quota resources = %+v", defaults)
	}

	fractional := resourcesToLinux(
		map[string]float64{"CPU": 1500},
		config.CPULimitModeQuota,
	)
	if fractional.CpuQuota != 150000 || fractional.CpuPeriod != 100000 {
		t.Fatalf("fractional quota resources = %+v", fractional)
	}

	minimum := resourcesToLinux(
		map[string]float64{"CPU": 1},
		config.CPULimitModeQuota,
	)
	if minimum.CpuQuota != 1000 {
		t.Fatalf("minimum quota = %d, want 1000", minimum.CpuQuota)
	}
}
