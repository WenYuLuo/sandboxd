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

package networkmanager

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// Local CPU count cache. It stays package-local so the network pool does not
// depend on cgroup-manager implementation details.
var (
	fetchCpu sync.Once
	cpuNum   int
	cpuError error
)

func getLocalCpuNum() (int, error) {
	if cpuNum > 0 {
		return cpuNum, nil
	}
	fetchCpu.Do(func() {
		quota, err := os.ReadFile("/sys/fs/cgroup/cpu/cpu.cfs_quota_us")
		if err != nil {
			cpuError = fmt.Errorf("read /sys/fs/cgroup/cpu/cpu.cfs_quota_us failed: %v", err)
			return
		}
		period, err := os.ReadFile("/sys/fs/cgroup/cpu/cpu.cfs_period_us")
		if err != nil {
			cpuError = fmt.Errorf("read /sys/fs/cgroup/cpu/cpu.cfs_period_us failed: %v", err)
			return
		}
		quotaInt, _ := strconv.Atoi(strings.Trim(string(quota), "\n"))
		periodInt, _ := strconv.Atoi(strings.Trim(string(period), "\n"))
		logrus.Debugf("local cpu quota is %v, period is %v", quotaInt, periodInt)

		if quotaInt == -1 {
			cpuNum, cpuError = getCpuCountFromCpuset()
			return
		}

		cpuNum = int(float64(quotaInt) / float64(periodInt))
	})
	return cpuNum, cpuError
}

func getCpuCountFromCpuset() (int, error) {
	data, err := os.ReadFile("/sys/fs/cgroup/cpuset/cpuset.cpus")
	if err != nil {
		return 0, fmt.Errorf("read /sys/fs/cgroup/cpuset/cpuset.cpus failed: %v", err)
	}

	cpus := strings.Trim(string(data), "\n")
	if cpus == "" {
		return 0, fmt.Errorf("cpuset.cpus is empty")
	}

	count := 0
	for _, r := range strings.Split(cpus, ",") {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if strings.Contains(r, "-") {
			parts := strings.Split(r, "-")
			if len(parts) != 2 {
				continue
			}
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 != nil || err2 != nil || start > end {
				continue
			}
			count += end - start + 1
		} else if _, err := strconv.Atoi(r); err == nil {
			count++
		}
	}
	return count, nil
}
