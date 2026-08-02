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
	"path/filepath"
	"testing"

	"github.com/inclusionAI/sandboxd/config"
	runscapi "github.com/inclusionAI/sandboxd/pkg/runtime/runsc"
	"github.com/stretchr/testify/assert"
)

func TestNewRunscHandlerUsesSharedLogFile(t *testing.T) {
	baseDir := t.TempDir()
	rootDir := filepath.Join(baseDir, "sandboxd", "root")
	cfg := config.Config{RootDir: rootDir}
	cfg.RuntimeConfig.FilestoreDir = filepath.Join(baseDir, "filestore")
	handler, err := NewRunscHandler(cfg, "/usr/local/bin/runsc", nil)
	assert.NoError(t, err)

	client, ok := handler.runsc.(*runscapi.Client)
	if !ok {
		t.Fatalf("runsc client has unexpected type %T", handler.runsc)
	}
	assert.Equal(t, filepath.Join(baseDir, "logs", config.RuntimeNameRunsc, "runsc.log"), client.Options.DebugLogPath)
}

func TestNewRunscHandlerPropagatesIgnoreCgroups(t *testing.T) {
	rootDir := filepath.Join(t.TempDir(), "sandboxd", "root")
	cfg := config.Config{RootDir: rootDir}
	cfg.DisableCgroup = true
	cfg.RuntimeConfig.FilestoreDir = filepath.Join(t.TempDir(), "filestore")
	handler, err := NewRunscHandler(cfg, "/usr/local/bin/runsc", nil)
	assert.NoError(t, err)

	client := handler.runsc.(*runscapi.Client)
	assert.True(t, client.Options.IgnoreCgroups)
}

func TestNewRunscHandlerRejectsMissingFilestore(t *testing.T) {
	rootDir := filepath.Join(t.TempDir(), "sandboxd", "root")
	_, err := NewRunscHandler(config.Config{RootDir: rootDir}, "/usr/local/bin/runsc", nil)
	assert.ErrorContains(t, err, "plugin.runtime.filestore_dir")
}
