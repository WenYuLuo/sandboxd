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

package imgcgroup

import (
	"os/exec"
	"testing"
)

func TestNilControllerEnabled(t *testing.T) {
	var c *Controller
	if c.Enabled() {
		t.Error("nil controller should not be enabled")
	}
}

func TestNilControllerApply(t *testing.T) {
	var c *Controller
	cmd := exec.Command("echo")
	// Should not panic
	c.Apply(cmd)
}

func TestNilControllerAddPID(t *testing.T) {
	var c *Controller
	if err := c.AddPID(1234); err != nil {
		t.Errorf("nil controller AddPID should be no-op, got: %v", err)
	}
}

func TestNilControllerClose(t *testing.T) {
	var c *Controller
	if err := c.Close(); err != nil {
		t.Errorf("nil controller Close should be no-op, got: %v", err)
	}
}

func TestControllerEnabled(t *testing.T) {
	c := &Controller{cgroupVersion: 1, cgroupDir: "/tmp/test", cgroupFD: -1}
	if !c.Enabled() {
		t.Error("non-nil controller should be enabled")
	}
}
