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

package bridge

import (
	"errors"
	"testing"
)

type iptablesCall struct {
	operation string
	table     string
	chain     string
	rulespec  []string
}

type fakeIPTablesClient struct {
	calls  []iptablesCall
	failAt int
}

func (f *fakeIPTablesClient) record(operation, table, chain string, rulespec []string) error {
	f.calls = append(f.calls, iptablesCall{operation, table, chain, append([]string(nil), rulespec...)})
	if f.failAt > 0 && len(f.calls) == f.failAt {
		return errors.New("injected iptables failure")
	}
	return nil
}

func (f *fakeIPTablesClient) Append(table, chain string, rulespec ...string) error {
	return f.record("append", table, chain, rulespec)
}

func (f *fakeIPTablesClient) AppendUnique(table, chain string, rulespec ...string) error {
	return f.record("append-unique", table, chain, rulespec)
}

func (f *fakeIPTablesClient) Delete(table, chain string, rulespec ...string) error {
	return f.record("delete", table, chain, rulespec)
}

func (f *fakeIPTablesClient) DeleteIfExists(table, chain string, rulespec ...string) error {
	return f.record("delete-if-exists", table, chain, rulespec)
}

func (f *fakeIPTablesClient) Exists(string, string, ...string) (bool, error) {
	return false, nil
}

func useFakeIPTables(t *testing.T, fake *fakeIPTablesClient) {
	t.Helper()
	original := newIPTablesClient
	newIPTablesClient = func() (iptablesClient, error) { return fake, nil }
	t.Cleanup(func() { newIPTablesClient = original })
}

func TestSetupDNATRuleHandlesIngressTraffic(t *testing.T) {
	fake := &fakeIPTablesClient{}
	useFakeIPTables(t, fake)

	if err := (BridgeNetworkManager{}).SetupDNATRule("tcp", 21008, "10.88.2.17", 50090); err != nil {
		t.Fatalf("SetupDNATRule() error = %v", err)
	}
	if len(fake.calls) != 2 {
		t.Fatalf("SetupDNATRule() calls = %d, want 2: %+v", len(fake.calls), fake.calls)
	}
	if fake.calls[0].table != "nat" || fake.calls[0].chain != "PREROUTING" {
		t.Fatalf("first call = %+v, want nat/PREROUTING", fake.calls[0])
	}
	if fake.calls[1].table != "filter" || fake.calls[1].chain != "FORWARD" {
		t.Fatalf("second call = %+v, want filter/FORWARD", fake.calls[1])
	}
}

func TestSetupDNATRuleRollsBackOnForwardFailure(t *testing.T) {
	fake := &fakeIPTablesClient{failAt: 2}
	useFakeIPTables(t, fake)

	err := (BridgeNetworkManager{}).SetupDNATRule("tcp", 21008, "10.88.2.17", 50090)
	if err == nil {
		t.Fatal("SetupDNATRule() succeeded, want failure")
	}
	if len(fake.calls) != 3 {
		t.Fatalf("SetupDNATRule() calls = %d, want 3: %+v", len(fake.calls), fake.calls)
	}
	if fake.calls[2].operation != "delete" || fake.calls[2].chain != "PREROUTING" {
		t.Fatalf("rollback = %+v, want PREROUTING delete", fake.calls[2])
	}
}

func TestSetupLocalDNATRuleHandlesLocalTraffic(t *testing.T) {
	fake := &fakeIPTablesClient{}
	useFakeIPTables(t, fake)

	if err := (BridgeNetworkManager{}).SetupLocalDNATRule("tcp", 21008, "10.88.2.17", 50090); err != nil {
		t.Fatalf("SetupLocalDNATRule() error = %v", err)
	}
	if len(fake.calls) != 1 {
		t.Fatalf("SetupLocalDNATRule() calls = %d, want 1: %+v", len(fake.calls), fake.calls)
	}
	wantPrefix := []string{"-p", "tcp", "-m", "addrtype", "--dst-type", "LOCAL"}
	for index, want := range wantPrefix {
		if fake.calls[0].rulespec[index] != want {
			t.Fatalf("OUTPUT rule = %v, want prefix %v", fake.calls[0].rulespec, wantPrefix)
		}
	}
}

func TestCleanupDNATRuleRemovesIngressRules(t *testing.T) {
	fake := &fakeIPTablesClient{}
	useFakeIPTables(t, fake)

	if err := (BridgeNetworkManager{}).CleanupDNATRule("tcp", 21008, "10.88.2.17", 50090); err != nil {
		t.Fatalf("CleanupDNATRule() error = %v", err)
	}
	if len(fake.calls) != 2 {
		t.Fatalf("CleanupDNATRule() calls = %d, want 2: %+v", len(fake.calls), fake.calls)
	}
	wantChains := []string{"PREROUTING", "FORWARD"}
	for index, want := range wantChains {
		if fake.calls[index].operation != "delete-if-exists" || fake.calls[index].chain != want {
			t.Fatalf("cleanup call %d = %+v, want %s delete", index, fake.calls[index], want)
		}
	}
}

func TestCleanupLocalDNATRuleRemovesLocalRule(t *testing.T) {
	fake := &fakeIPTablesClient{}
	useFakeIPTables(t, fake)

	if err := (BridgeNetworkManager{}).CleanupLocalDNATRule("tcp", 21008, "10.88.2.17", 50090); err != nil {
		t.Fatalf("CleanupLocalDNATRule() error = %v", err)
	}
	if len(fake.calls) != 1 || fake.calls[0].operation != "delete-if-exists" || fake.calls[0].chain != "OUTPUT" {
		t.Fatalf("cleanup call = %+v, want OUTPUT delete", fake.calls)
	}
}
