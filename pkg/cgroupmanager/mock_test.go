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

package cgroupmanager

import (
	cg "github.com/containerd/cgroups/v3/cgroup1"
	"github.com/containerd/cgroups/v3/cgroup1/stats"
	spec "github.com/opencontainers/runtime-spec/specs-go"
)

// MockCgroup mirrors the no-op cg.Cgroup that used to live in the deleted
// pkg/resourcemanager/manager_test.go. It exists only to satisfy
// FakeCgroupHandler in cgroup_extend_test.go.
type MockCgroup struct {
	path string
}

func (m *MockCgroup) New(s string, resources *spec.LinuxResources) (cg.Cgroup, error) {
	return &MockCgroup{path: s}, nil
}
func (m *MockCgroup) Add(process cg.Process, name ...cg.Name) error     { return nil }
func (m *MockCgroup) AddProc(u uint64, name ...cg.Name) error           { return nil }
func (m *MockCgroup) AddTask(process cg.Process, name ...cg.Name) error { return nil }
func (m *MockCgroup) Delete() error                                     { return nil }
func (m *MockCgroup) MoveTo(cgroup cg.Cgroup) error                     { return nil }
func (m *MockCgroup) Stat(handler ...cg.ErrorHandler) (*stats.Metrics, error) {
	return nil, nil
}
func (m *MockCgroup) Update(resources *spec.LinuxResources) error               { return nil }
func (m *MockCgroup) Processes(name cg.Name, b bool) ([]cg.Process, error)      { return nil, nil }
func (m *MockCgroup) Tasks(name cg.Name, b bool) ([]cg.Task, error)             { return nil, nil }
func (m *MockCgroup) Freeze() error                                             { return nil }
func (m *MockCgroup) Thaw() error                                               { return nil }
func (m *MockCgroup) OOMEventFD() (uintptr, error)                              { return 0, nil }
func (m *MockCgroup) RegisterMemoryEvent(event cg.MemoryEvent) (uintptr, error) { return 0, nil }
func (m *MockCgroup) State() cg.State                                           { return cg.Unknown }
func (m *MockCgroup) Subsystems() []cg.Subsystem                                { return []cg.Subsystem{} }

var _ cg.Cgroup = &MockCgroup{}
