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
	"testing"
)

func TestSplitObject(t *testing.T) {
	cases := []struct {
		name       string
		object     string
		wantPrefix string
		wantName   string
		wantErr    bool
	}{
		{name: "single-segment", object: "foo", wantPrefix: "", wantName: "foo"},
		{name: "nested", object: "dir/sub/file", wantPrefix: "dir/sub/", wantName: "file"},
		{name: "leading-slash", object: "/dir/file", wantPrefix: "/dir/", wantName: "file"},
		{name: "dot-clean", object: "dir/./file", wantPrefix: "dir/", wantName: "file"},
		{name: "trailing-slash", object: "dir/", wantErr: true},
		{name: "empty", object: "", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			prefix, name, err := splitObject(tc.object)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if prefix != tc.wantPrefix || name != tc.wantName {
				t.Fatalf("got prefix=%q name=%q, want prefix=%q name=%q", prefix, name, tc.wantPrefix, tc.wantName)
			}
		})
	}
}

const testImageURL = "registry.example.com/akernel/client:f6e3adaf-20260128200059"

func TestOCIMountRequestString(t *testing.T) {
	req := OCIMountRequest{ImageURL: testImageURL}
	got := req.String()
	want := "(" + testImageURL + ")"
	if got != want {
		t.Errorf("OCIMountRequest.String() = %q, want %q", got, want)
	}
}

func TestOCIUmountRequestString(t *testing.T) {
	req := OCIUmountRequest{ImageURL: testImageURL}
	got := req.String()
	want := "(" + testImageURL + ")"
	if got != want {
		t.Errorf("OCIUmountRequest.String() = %q, want %q", got, want)
	}
}
