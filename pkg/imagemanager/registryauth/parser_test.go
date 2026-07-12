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

package registryauth

import "testing"

func TestParseFlatFormat(t *testing.T) {
	data := []byte(`{
		"registry.example.com": {"Auth":"host-auth"},
		"registry.example.com/demo/image": {"Auth":"repo-auth"}
	}`)

	auths, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if got := auths["registry.example.com"].Auth; got != "host-auth" {
		t.Fatalf("auths[host].Auth = %q, want %q", got, "host-auth")
	}
	if got := auths["registry.example.com/demo/image"].Auth; got != "repo-auth" {
		t.Fatalf("auths[host/repo].Auth = %q, want %q", got, "repo-auth")
	}
}

func TestParseDockerAuthsFormat(t *testing.T) {
	data := []byte(`{
		"auths": {
			"registry.example.com": {"auth":"host-auth"},
			"registry.example.com/demo/image": {"Auth":"repo-auth"}
		},
		"credsStore": "desktop"
	}`)

	auths, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if got := auths["registry.example.com"].Auth; got != "host-auth" {
		t.Fatalf("auths[host].Auth = %q, want %q", got, "host-auth")
	}
	if got := auths["registry.example.com/demo/image"].Auth; got != "repo-auth" {
		t.Fatalf("auths[host/repo].Auth = %q, want %q", got, "repo-auth")
	}
}

func TestParseDockerURLKeyNormalization(t *testing.T) {
	data := []byte(`{
		"auths": {
			"https://registry.example.com/v1/": {"auth":"host-auth"},
			"https://registry.example.com/demo/image/": {"auth":"repo-auth"}
		}
	}`)

	auths, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if got := auths["registry.example.com"].Auth; got != "host-auth" {
		t.Fatalf("auths[normalized host].Auth = %q, want %q", got, "host-auth")
	}
	if got := auths["registry.example.com/demo/image"].Auth; got != "repo-auth" {
		t.Fatalf("auths[normalized repo].Auth = %q, want %q", got, "repo-auth")
	}
}
