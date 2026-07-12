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

package util

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os/exec"
	"strings"
)

type CmdExecutor interface {
	Execute(ctx context.Context, cmd string, args ...string) ([]byte, error)
}

type SystemCmdExecutor struct{}

func (s *SystemCmdExecutor) Execute(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	output, err := exec.CommandContext(ctx, cmd, args[0:]...).CombinedOutput()
	logrus.Debugf("executed cmd: %s %s, output: %s, err: %v", cmd, args, string(output), err)
	return output, err
}

type MockCmdExecutor struct {
	// if the cmd is included in this map, return success
	SuccessMap map[string]bool
	OutputMap  map[string]string
}

func (m *MockCmdExecutor) Execute(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	if m.getSuccess(strings.Join(append([]string{cmd}, args...), " ")) {
		return m.getOutput(strings.Join(append([]string{cmd}, args...), " ")), nil
	} else {
		return m.getOutput(strings.Join(append([]string{cmd}, args...), " ")), fmt.Errorf("mock error")
	}
}

func (m *MockCmdExecutor) getSuccess(cmd string) bool {
	if m.SuccessMap == nil {
		return false
	}
	for key, success := range m.SuccessMap {
		if strings.Contains(cmd, key) {
			return success
		}
	}
	return false
}

func (m *MockCmdExecutor) getOutput(cmd string) []byte {
	if m.OutputMap == nil {
		return []byte("")
	}
	for key, output := range m.OutputMap {
		if strings.Contains(cmd, key) {
			return []byte(output)
		}
	}
	return []byte("")
}
