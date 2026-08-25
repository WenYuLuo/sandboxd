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
	"net"
	"testing"

	"github.com/inclusionAI/sandboxd/config"
	"github.com/inclusionAI/sandboxd/pkg/networkmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequiredStartResourcesWithoutCgroup(t *testing.T) {
	resources, err := requiredStartResources(config.RuntimeNameRunsc, true)
	require.NoError(t, err)
	assert.Equal(t, []string{config.ResourceNameInterface}, resources)
}

func TestRequiredStartResourcesWithCgroup(t *testing.T) {
	resources, err := requiredStartResources(config.RuntimeNameRunsc, false)
	require.NoError(t, err)
	assert.Equal(t, config.RuntimeResources[config.RuntimeNameRunsc], resources)
}

func TestFirecrackerRequiresCgroup(t *testing.T) {
	_, err := requiredStartResources(config.RuntimeNameFirecracker, true)
	require.ErrorContains(t, err, "requires cgroup management")

	resources, err := requiredStartResources(config.RuntimeNameFirecracker, false)
	require.NoError(t, err)
	assert.Equal(t, config.RuntimeResources[config.RuntimeNameFirecracker], resources)
}

func TestStartSuccessResponseIncludesCommittedEndpoint(t *testing.T) {
	ports := []string{"tcp:18080:8080"}
	response := newStartSuccessResponse("sandbox-1", ports, &networkmanager.NetResource{
		Ip:                 net.ParseIP("10.88.0.2"),
		EndpointGeneration: 42,
	})

	assert.Equal(t, int32(0), response.Code)
	assert.Equal(t, "sandbox-1", response.ID)
	assert.Equal(t, []string{"tcp:18080:8080"}, response.Ports)
	assert.Equal(t, "10.88.0.2", response.BridgeIp)
	assert.Equal(t, uint64(42), response.EndpointGeneration)

	ports[0] = "tcp:19090:9090"
	assert.Equal(t, []string{"tcp:18080:8080"}, response.Ports)
}
