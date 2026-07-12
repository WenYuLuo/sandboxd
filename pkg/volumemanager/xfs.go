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

// Package volumemanager owns the optional reflink-capable XFS image used as
// overlay backing storage.
package volumemanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsXFSMounted reports whether dir is currently mounted as an XFS filesystem.
// It lives outside the runsc package so lifecycle code can share the probe.
func IsXFSMounted(dir string) (bool, error) {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return false, fmt.Errorf("read /proc/mounts: %v", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == dir && fields[2] == "xfs" {
			return true, nil
		}
	}
	return false, nil
}

// EnsureXFSMount ensures filestoreDir is backed by an XFS filesystem.
//
// If <parent>/xfs.img does not exist, a new image of the given size is
// created and formatted using mkfs.xfs (reflink=1), then mounted via a
// loop device. mkfs.xfs creates the image file itself; truncate sets the
// size up front. If xfs.img already exists (e.g. after a crash), it is
// remounted as-is provided it is not already mounted.
//
// Behavior is unchanged from the original runsc_handler.go implementation;
// only the package home moved.
func EnsureXFSMount(filestoreDir, size string) error {
	if err := os.MkdirAll(filestoreDir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %v", filestoreDir, err)
	}

	imgFile := filepath.Join(filepath.Dir(filestoreDir), "xfs.img")

	mounted, err := IsXFSMounted(filestoreDir)
	if err != nil {
		return err
	}
	if mounted {
		return nil
	}

	os.Remove(imgFile)

	if out, err := exec.Command("truncate", "-s", size, imgFile).CombinedOutput(); err != nil {
		return fmt.Errorf("truncate %s: %s: %v", imgFile, out, err)
	}
	out, err := exec.Command("mkfs.xfs", "-f", "-m", "reflink=1", "-i", "nrext64=0", imgFile).CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "unknown option") && strings.Contains(string(out), "nrext64") {
			out, err = exec.Command("mkfs.xfs", "-f", "-m", "reflink=1", imgFile).CombinedOutput()
			if err != nil {
				os.Remove(imgFile)
				return fmt.Errorf("mkfs.xfs: %s: %v", out, err)
			}
		} else {
			os.Remove(imgFile)
			return fmt.Errorf("mkfs.xfs: %s: %v", out, err)
		}
	}

	if out, err := exec.Command("mount", "-o", "loop,defaults,discard", imgFile, filestoreDir).CombinedOutput(); err != nil {
		os.Remove(imgFile)
		return fmt.Errorf("mount xfs %s -> %s: %s: %v", imgFile, filestoreDir, out, err)
	}

	return nil
}

// CleanupXFSMount unmounts the XFS filesystem at filestoreDir (if mounted),
// then removes the image file and the mount point directory. No-op when
// filestoreDir is empty or not currently mounted as XFS.
func CleanupXFSMount(filestoreDir string) error {
	if filestoreDir == "" {
		return nil
	}
	mounted, err := IsXFSMounted(filestoreDir)
	if err != nil {
		return err
	}
	if !mounted {
		return nil
	}
	if out, err := exec.Command("umount", filestoreDir).CombinedOutput(); err != nil {
		return fmt.Errorf("umount %s: %s: %v", filestoreDir, out, err)
	}
	imgFile := filepath.Join(filepath.Dir(filestoreDir), "xfs.img")
	os.Remove(imgFile)
	os.Remove(filestoreDir)
	return nil
}
