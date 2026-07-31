// Copyright (c) 2026 Ant Group Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runsc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateUsesExactDebugLogPath(t *testing.T) {
	tempDir := t.TempDir()
	argsFile := filepath.Join(tempDir, "args")
	binary := filepath.Join(tempDir, "runsc")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$RUNSC_TEST_ARGS\"\n"
	if err := os.WriteFile(binary, []byte(script), 0755); err != nil {
		t.Fatalf("write fake runsc: %v", err)
	}
	t.Setenv("RUNSC_TEST_ARGS", argsFile)

	debugLogPath := filepath.Join(tempDir, "logs", "runsc.log")
	client := NewClientWithOptions(binary, filepath.Join(tempDir, "root"), Options{
		DebugLogPath: debugLogPath,
	})
	if err := client.Create(context.Background(), StartArgs{
		ID:         "sandbox-id",
		BundleDir:  filepath.Join(tempDir, "bundle"),
		UserStdout: filepath.Join(tempDir, "stdout"),
		UserStderr: filepath.Join(tempDir, "stderr"),
	}); err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("read fake runsc arguments: %v", err)
	}
	args := strings.Split(strings.TrimSpace(string(data)), "\n")
	want := "-debug-log=" + debugLogPath
	for _, arg := range args {
		if arg == want {
			return
		}
	}
	t.Fatalf("runsc arguments %q do not contain %q", args, want)
}

func TestGlobalArgsIgnoreCgroups(t *testing.T) {
	client := NewClientWithOptions("/usr/local/bin/runsc", "/run/runsc", Options{
		IgnoreCgroups: true,
	})
	args := client.globalArgs()
	if got := strings.Join(args, " "); got != "--root /run/runsc --ignore-cgroups" {
		t.Fatalf("global args = %q", got)
	}
}
