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

package networkacl

import (
	"fmt"

	"github.com/inclusionAI/sandboxd/internal/diagnostics"
	"github.com/inclusionAI/sandboxd/pkg/errord"
)

type removal struct {
	done chan struct{}
	err  error
}

// Remove reserves the endpoint until both kernel cleanup and its final durable
// removal succeed. Other endpoints may make progress during slow iptables and
// conntrack operations. Restore/Close wait for all such operations to settle.
func (m *Manager) Remove(sandboxID string) (result error) {
	m.lifecycleMu.RLock()
	defer m.lifecycleMu.RUnlock()
	unlockPrepare := diagnostics.Lock(&m.mu, "acl.remove.prepare")
	if m.closed {
		unlockPrepare()
		return errord.ErrUnavailable
	}
	if pending := m.removals[sandboxID]; pending != nil {
		unlockPrepare()
		<-pending.done
		// Join the original operation, not whatever incarnation may now own the ID.
		return pending.err
	}
	entry, ok := m.entries[sandboxID]
	if !ok {
		unlockPrepare()
		return nil
	}
	if m.iptables == nil {
		defer unlockPrepare()
		return m.reconcileOrphansLocked(map[string]persistedEntry{sandboxID: entry})
	}
	if err := m.checkCleanupOwnershipLocked(sandboxID, entry); err != nil {
		unlockPrepare()
		return err
	}
	if !entry.Orphaned {
		previous := entry
		sourceOwner, sourcePresent := m.sourceIndex[entry.IP]
		entry.Orphaned = true
		m.entries[sandboxID] = entry
		if sourceOwner == sandboxID {
			delete(m.sourceIndex, entry.IP)
		}
		if err := m.persistLocked(); err != nil {
			m.entries[sandboxID] = previous
			if sourcePresent {
				m.sourceIndex[entry.IP] = sourceOwner
			}
			unlockPrepare()
			return fmt.Errorf("persist network ACL cleanup intent: %w", err)
		}
	}
	if m.removals == nil {
		m.removals = make(map[string]*removal)
	}
	pending := &removal{done: make(chan struct{})}
	m.removals[sandboxID] = pending
	unlockPrepare()

	cleanupErr := m.iptables.cleanup(entry)

	defer diagnostics.Lock(&m.mu, "acl.remove.commit")()
	defer func() { pending.err = result; delete(m.removals, sandboxID); close(pending.done) }()
	if cleanupErr != nil {
		return fmt.Errorf("clean network ACL for %s: %w", sandboxID, cleanupErr)
	}
	delete(m.entries, sandboxID)
	if err := m.persistLocked(); err != nil {
		m.entries[sandboxID] = entry
		return fmt.Errorf("persist removal of cleaned network ACL state: %w", err)
	}
	return nil
}

func (m *Manager) checkCleanupOwnershipLocked(sandboxID string, entry persistedEntry) error {
	for otherID, other := range m.entries {
		if otherID == sandboxID {
			continue
		}
		conflicts := other.IfIndex == entry.IfIndex || other.IP == entry.IP || other.HostVeth == entry.HostVeth
		if conflicts && (!other.Orphaned || m.removals[otherID] != nil) {
			return fmt.Errorf("refusing cleanup of ifindex %d owned by active sandbox %s", entry.IfIndex, otherID)
		}
	}
	return nil
}
