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

// Package cgroup manages a memory cgroup for distill_fs daemon processes.
//
// On Linux, it creates a /distillfs cgroup (v1 or v2) and places daemon
// processes into it. On other platforms, all operations are no-ops.
// A nil *Controller is safe to use — all methods become no-ops.
package imgcgroup

import "os/exec"

// Controller manages a cgroup for distill_fs daemons.
// A nil *Controller is safe to use; all methods become no-ops.
type Controller struct {
	cgroupVersion int    // 1 or 2
	cgroupDir     string // path to the cgroup directory
	cgroupFD      int    // file descriptor for cgroup dir (v2 only, -1 if unused)
}

// Enabled reports whether the controller is active.
func (c *Controller) Enabled() bool {
	return c != nil
}

// Apply configures cmd to start directly in the cgroup (v2) or is a no-op (v1/nil).
// Must be called before cmd.Start().
func (c *Controller) Apply(cmd *exec.Cmd) {
	if c == nil {
		return
	}
	c.applyPlatform(cmd)
}

// AddPID writes a process ID to cgroup.procs.
// On v2 this is a no-op (processes are placed via Apply).
// On v1 this is the primary mechanism, called after the daemon is confirmed running.
func (c *Controller) AddPID(pid int) error {
	if c == nil {
		return nil
	}
	return c.addPIDPlatform(pid)
}

// Close releases resources held by the controller (e.g. cgroup dir fd).
func (c *Controller) Close() error {
	if c == nil {
		return nil
	}
	return c.closePlatform()
}
