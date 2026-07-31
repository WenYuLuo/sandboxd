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

package xpumanager

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	api "github.com/inclusionAI/sandboxd/api/runtime/v1"
	svc "github.com/inclusionAI/sandboxd/pkg/runtime"
)

const (
	nvidiaVisibleDevicesEnv  = "NVIDIA_VISIBLE_DEVICES"
	nvidiaDriverCapabilities = "NVIDIA_DRIVER_CAPABILITIES"
	cudaVisibleDevicesEnv    = "CUDA_VISIBLE_DEVICES"
	nvidiaDriverCapsValue    = "compute,utility"
	nvidiaRuntimeHookPath    = "/usr/bin/nvidia-container-runtime-hook"
	nvidiaRuntimeHookArg     = "nvidia-container-runtime-hook"
	nvidiaContainerCLI       = "nvidia-container-cli"
	nvidiaControlDevice      = "/dev/nvidiactl"
	nvidiaUVMDevice          = "/dev/nvidia-uvm"
	nvidiaDiscoveryTimeout   = 15 * time.Second
)

var nvidiaModelSeparator = regexp.MustCompile(`[^a-z0-9._-]+`)

type commandRunner func(context.Context, string, ...string) ([]byte, error)
type statFunc func(string) (os.FileInfo, error)

func runCommand(ctx context.Context, binary string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, binary, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf(
			"%s %s: %w: %s",
			binary,
			strings.Join(args, " "),
			err,
			strings.TrimSpace(string(output)),
		)
	}
	return output, nil
}

func (m *Manager) discoverNVIDIA() error {
	if m.runscBinary == "" {
		return errors.New("runsc runtime is not configured")
	}
	cliPath, err := exec.LookPath(nvidiaContainerCLI)
	if err != nil {
		return fmt.Errorf("locate %s: %w", nvidiaContainerCLI, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), nvidiaDiscoveryTimeout)
	defer cancel()
	infoOutput, err := m.run(ctx, cliPath, "--load-kmods", "info")
	if err != nil {
		return fmt.Errorf("discover NVIDIA devices: %w", err)
	}
	driverVersion, devices, err := parseNVIDIAInfo(string(infoOutput))
	if err != nil {
		return err
	}
	supportedOutput, err := m.run(ctx, m.runscBinary, "nvproxy", "list-supported-drivers")
	if err != nil {
		return fmt.Errorf("list runsc nvproxy drivers: %w", err)
	}
	if !nvidiaDriverSupported(driverVersion, string(supportedOutput)) {
		return fmt.Errorf(
			"NVIDIA driver %s is not supported by %s nvproxy",
			driverVersion,
			m.runscBinary,
		)
	}
	for _, path := range []string{nvidiaControlDevice, nvidiaUVMDevice} {
		if _, err := m.stat(path); err != nil {
			return fmt.Errorf("required NVIDIA device %s is unavailable: %w", path, err)
		}
	}

	m.devices = devices
	m.resources = buildResources(devices)
	return nil
}

// ReservedEnv reports whether key is controlled by the NVIDIA provider.
func ReservedEnv(key string) bool {
	switch key {
	case nvidiaVisibleDevicesEnv, nvidiaDriverCapabilities, cudaVisibleDevicesEnv:
		return true
	default:
		return false
	}
}

func nvidiaSpecUpdates(uuids []string, recordJSON []byte) *svc.SpecUpdates {
	logicalIDs := make([]string, len(uuids))
	for index := range uuids {
		logicalIDs[index] = strconv.Itoa(index)
	}
	return &svc.SpecUpdates{
		Envs: []*api.KeyValue{
			{Key: nvidiaVisibleDevicesEnv, Value: strings.Join(uuids, ",")},
			{Key: nvidiaDriverCapabilities, Value: nvidiaDriverCapsValue},
			{Key: cudaVisibleDevicesEnv, Value: strings.Join(logicalIDs, ",")},
		},
		Prestart: []svc.Hook{{
			Path: nvidiaRuntimeHookPath,
			Args: []string{nvidiaRuntimeHookArg, "prestart"},
		}},
		Annotations: map[string]string{
			AllocationAnnotation: string(recordJSON),
		},
		RequiresHostWritableRootfs: true,
	}
}

func parseNVIDIAInfo(output string) (string, map[uint32]Device, error) {
	driverVersion := ""
	devices := make(map[uint32]Device)
	var current *Device
	commit := func() error {
		if current == nil {
			return nil
		}
		if current.UUID == "" || current.ProductModel == "" {
			return fmt.Errorf("incomplete NVIDIA device record for index %d", current.ID)
		}
		if _, duplicate := devices[current.ID]; duplicate {
			return fmt.Errorf("duplicate NVIDIA device index %d", current.ID)
		}
		devices[current.ID] = *current
		current = nil
		return nil
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "NVRM version":
			driverVersion = value
		case "Device Index":
			if err := commit(); err != nil {
				return "", nil, err
			}
			parsed, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				return "", nil, fmt.Errorf("parse NVIDIA device index %q: %w", value, err)
			}
			current = &Device{ID: uint32(parsed)}
		case "Model":
			if current != nil {
				current.ProductModel = normalizeNVIDIAModel(value)
			}
		case "GPU UUID":
			if current != nil {
				current.UUID = value
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", nil, err
	}
	if err := commit(); err != nil {
		return "", nil, err
	}
	if driverVersion == "" {
		return "", nil, errors.New("NVIDIA discovery output has no NVRM version")
	}
	if len(devices) == 0 {
		return "", nil, errors.New("NVIDIA discovery output has no GPU devices")
	}
	return driverVersion, devices, nil
}

func nvidiaDriverSupported(driverVersion, output string) bool {
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == driverVersion {
			return true
		}
	}
	return false
}

func normalizeNVIDIAModel(model string) string {
	model = strings.ToLower(strings.TrimSpace(model))
	model = strings.TrimSpace(strings.TrimPrefix(model, "nvidia "))
	model = nvidiaModelSeparator.ReplaceAllString(model, "-")
	return strings.Trim(model, "-._")
}
