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

//go:build linux

package oci

import "golang.org/x/sys/unix"

func mountReadonlyOverlay(target, mountOpts string) error {
	// Readonly is guaranteed by lowerdir-only overlay; passing MS_RDONLY can
	// trigger EINVAL on some kernels for overlay mounts.
	return unix.Mount("overlay", target, "overlay", 0, mountOpts)
}

func mountReadonlyBind(source, target string) error {
	if err := unix.Mount(source, target, "", uintptr(unix.MS_BIND), ""); err != nil {
		return err
	}
	return unix.Mount("", target, "", uintptr(unix.MS_BIND|unix.MS_REMOUNT|unix.MS_RDONLY), "")
}

func unmountOverlay(target string) error {
	return unix.Unmount(target, 0)
}
