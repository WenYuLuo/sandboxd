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

package util

import (
	"fmt"
	"path/filepath"
	"strings"
)

// JoinWithinRoot joins path elements while ensuring the result remains below
// root. It is a defense-in-depth guard for paths derived from sandbox IDs.
func JoinWithinRoot(root string, elements ...string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve root %q: %w", root, err)
	}
	targetAbs, err := filepath.Abs(filepath.Join(append([]string{rootAbs}, elements...)...))
	if err != nil {
		return "", fmt.Errorf("resolve path below %q: %w", root, err)
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return "", fmt.Errorf("compare path %q with root %q: %w", targetAbs, rootAbs, err)
	}
	if rel == "." || rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes root %q", targetAbs, rootAbs)
	}
	return targetAbs, nil
}
