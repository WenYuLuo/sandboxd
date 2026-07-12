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

package cgroupops

import (
	cg "github.com/containerd/cgroups/v3/cgroup1"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type CgroupHandler interface {
	Create(path cg.Path, resources *specs.LinuxResources, opts ...cg.InitOpts) (cg.Cgroup, error)
	Load(path cg.Path, opts ...cg.InitOpts) (cg.Cgroup, error)
}

type CgroupHandlerImpl struct{}

func (h *CgroupHandlerImpl) Create(path cg.Path, resources *specs.LinuxResources, opts ...cg.InitOpts) (cg.Cgroup, error) {
	return cg.New(path, resources, opts...)
}

func (h *CgroupHandlerImpl) Load(path cg.Path, opts ...cg.InitOpts) (cg.Cgroup, error) {
	return cg.Load(path, opts...)
}
