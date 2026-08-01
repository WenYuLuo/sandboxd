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

import "testing"

func TestEphemeralStorageCapacity(t *testing.T) {
	m := NewModule(t.TempDir(), "")
	capacity, allocatable, err := m.EphemeralStorageCapacity()
	if err != nil {
		t.Fatal(err)
	}
	if capacity == 0 || allocatable == 0 || allocatable > capacity {
		t.Fatalf("capacity=%d allocatable=%d", capacity, allocatable)
	}
}

func TestEphemeralStorageCapacityRequiresFilestore(t *testing.T) {
	m := NewModule("", "")
	if _, _, err := m.EphemeralStorageCapacity(); err == nil {
		t.Fatal("expected unconfigured filestore error")
	}
}
