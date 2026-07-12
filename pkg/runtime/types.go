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

package runtime

import (
	"time"
)

const (
	defaultRunscRoot = "/run/containerd/runsc/default"
)

// UnionSandboxState is the state of a container returned by the list command
type UnionSandboxState struct {
	// ID is the container ID
	ID string `json:"id"`
	// InitProcessPid is the init process id in the parent namespace
	InitProcessPid int `json:"pid"`
	// Status is the current status of the container
	Status SandboxStatus `json:"status"`
	// Bundle is the path on the filesystem to the bundle
	Bundle string `json:"bundle"`
	// Created is the unix timestamp for the creation time of the container in UTC
	Created string `json:"created"`

	updateTime *time.Time
}

type SandboxStatus string

const (
	// SandboxStatusCreated is the status of a container after it has been created.
	SandboxStatusCreated SandboxStatus = "created"
	// SandboxStatusRunning is the status of a container after it has been started.
	SandboxStatusRunning SandboxStatus = "running"
	// SandboxStatusExited is the status of a container after it has exited.
	SandboxStatusExited SandboxStatus = "stopped"
	// SandboxStatusUnknown is the status of a container when its status cannot be determined.
	SandboxStatusUnknown SandboxStatus = "unknown"
)

type Exit struct {
	Timestamp time.Time
	Status    int
}
