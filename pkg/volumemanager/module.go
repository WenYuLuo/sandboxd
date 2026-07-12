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

package volumemanager

import (
	"os"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

// Module owns the optional XFS filestore lifecycle. It exposes:
//
//   - Start: probes / creates the XFS image at FilestoreDir; on failure
//     leaves the directory in place (overlay can still use it as a plain
//     tmpfs fallback).
//   - Stop: tears the XFS mount down via CleanupXFSMount.
//
// Module is safe to construct with FilestoreDir == "" — Start becomes a
// no-op, matching the behaviour when sandboxd runs without an XFS-backed
// filestore.
type Module struct {
	FilestoreDir string
	Size         string

	started    atomic.Bool
	xfsMounted atomic.Bool
}

// NewModule constructs a Module rooted at filestoreDir. size accepts any
// string that mkfs.xfs / truncate take (e.g. "100G").
func NewModule(filestoreDir, size string) *Module {
	return &Module{FilestoreDir: filestoreDir, Size: size}
}

// Start mounts the XFS filestore. When the loop device is unavailable or
// mkfs.xfs is missing the call falls back: the directory is created plain
// so overlay can still work.
func (m *Module) Start() error {
	m.started.Store(true)
	if m.FilestoreDir == "" {
		return nil
	}
	if m.Size == "" {
		return errMissingSize
	}
	if err := EnsureXFSMount(m.FilestoreDir, m.Size); err != nil {
		logrus.Warnf("volumemanager: XFS filestore mount failed (%v); falling back to a plain directory", err)
		// Best-effort: keep the directory so overlay can still use it as a
		// plain mount.
		_ = os.MkdirAll(m.FilestoreDir, 0755)
		return nil
	}
	m.xfsMounted.Store(true)
	return nil
}

// Stop unmounts the filestore (best-effort). Safe to call when Start was
// never called or when the mount never came up.
func (m *Module) Stop() error {
	if !m.started.Load() {
		return nil
	}
	return CleanupXFSMount(m.FilestoreDir)
}

// Healthy reports whether a started VolumeManager has its expected backing.
// An unconfigured filestore is treated as healthy because the operator has
// explicitly chosen not to use an XFS-backed filestore.
func (m *Module) Healthy() bool {
	if m.FilestoreDir == "" {
		return true
	}
	return m.xfsMounted.Load()
}

// errMissingSize is a typed sentinel so callers / tests can distinguish a
// config validation failure from a runtime mount failure.
var errMissingSize = errMissingSizeT{}

type errMissingSizeT struct{}

func (errMissingSizeT) Error() string {
	return "filestore_dir_size must be set when filestore_dir is configured"
}
