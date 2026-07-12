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

package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMountStore_PutGetDelete(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := OpenMountStore(dbPath)
	if err != nil {
		t.Fatalf("OpenMountStore: %v", err)
	}
	defer store.Close()

	record := &MountRecord{
		ImageURL:      "docker.io/library/alpine:latest",
		MountType:     "nydus",
		NydusImageURL: "docker.io/library/alpine:latest-nydus",
		MountPoint:    "/mnt/abc",
	}

	// Put
	if err := store.Put(record.ImageURL, record); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Get
	got, err := store.Get(record.ImageURL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got == nil {
		t.Fatal("Get returned nil for existing key")
	}
	if got.MountType != "nydus" {
		t.Errorf("MountType = %q, want %q", got.MountType, "nydus")
	}
	if got.NydusImageURL != record.NydusImageURL {
		t.Errorf("NydusImageURL = %q, want %q", got.NydusImageURL, record.NydusImageURL)
	}
	if got.MountPoint != record.MountPoint {
		t.Errorf("MountPoint = %q, want %q", got.MountPoint, record.MountPoint)
	}

	// Delete
	if err := store.Delete(record.ImageURL); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	got, err = store.Get(record.ImageURL)
	if err != nil {
		t.Fatalf("Get after delete: %v", err)
	}
	if got != nil {
		t.Error("Get returned non-nil after delete")
	}
}

func TestMountStore_GetMissing(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := OpenMountStore(dbPath)
	if err != nil {
		t.Fatalf("OpenMountStore: %v", err)
	}
	defer store.Close()

	got, err := store.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != nil {
		t.Error("expected nil for missing key")
	}
}

func TestMountStore_Reopen(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := OpenMountStore(dbPath)
	if err != nil {
		t.Fatalf("OpenMountStore: %v", err)
	}

	record := &MountRecord{
		ImageURL:  "myimage:v1",
		MountType: "oci",
	}
	if err := store.Put(record.ImageURL, record); err != nil {
		t.Fatalf("Put: %v", err)
	}
	store.Close()

	// Reopen and verify persistence
	store2, err := OpenMountStore(dbPath)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer store2.Close()

	got, err := store2.Get("myimage:v1")
	if err != nil {
		t.Fatalf("Get after reopen: %v", err)
	}
	if got == nil || got.MountType != "oci" {
		t.Errorf("expected oci record after reopen, got %+v", got)
	}
}

func TestMountStore_OpenInvalidPath(t *testing.T) {
	_, err := OpenMountStore(filepath.Join(os.DevNull, "impossible", "path.db"))
	if err == nil {
		t.Error("expected error for invalid path")
	}
}
