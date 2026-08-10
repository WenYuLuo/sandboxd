// Copyright (c) 2026 Ant Group Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package networkacl

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPTablesRulesAreStatefulAndDenyFirst(t *testing.T) {
	backend := &iptablesBackend{bridgeIP: net.ParseIP("10.88.0.1")}
	policy := Policy{
		DNS: &DNSPolicy{},
		Traffic: &TrafficPolicy{
			DefaultAction: actionDeny,
			Mode:          policyModeStateful,
			Rules: []TrafficRule{
				{
					Action: actionAllow, Directions: []uint8{directionIngress},
					Protocol: 6, PeerAny: true, SandboxPort: 50090,
				},
				{
					Action: actionDeny, Directions: []uint8{directionIngress},
					Protocol: 6, PeerIP: [4]byte{192, 0, 2, 10}, PeerPort: 32000,
					SandboxPort: 50090,
				},
			},
		},
	}
	rules := backend.compileRules(policy, directionIngress, 7)
	require.GreaterOrEqual(t, len(rules), 10)
	mark := "0xa5c10007"
	assert.Equal(t, []string{"-p", "tcp", "-s", "10.88.0.1", "--sport", "53", "-j", "RETURN"}, rules[0])
	assert.Equal(t, []string{"-m", "conntrack", "--ctstate", "INVALID", "-j", "DROP"}, rules[4])
	assert.Equal(t, []string{
		"-m", "conntrack", "--ctstate", "ESTABLISHED,RELATED",
		"-m", "connmark", "--mark", mark, "-j", "RETURN",
	}, rules[5])
	assert.Equal(t, []string{
		"-p", "tcp", "-s", "192.0.2.10", "--sport", "32000", "--dport", "50090", "-j", "DROP",
	}, rules[6])
	assert.Equal(t, []string{
		"-p", "tcp", "--dport", "50090", "-m", "conntrack", "--ctstate", "NEW",
		"-j", "CONNMARK", "--set-xmark", mark + "/0xffffffff",
	}, rules[7])
	assert.Equal(t, []string{"-m", "connmark", "--mark", mark, "-j", "RETURN"}, rules[8])
	assert.Equal(t, []string{"-j", "DROP"}, rules[9])
}

func TestIPTablesConnectionMarksChangeWithPolicyGeneration(t *testing.T) {
	assert.NotEqual(t, iptablesConnectionMark(1), iptablesConnectionMark(2))
	assert.NotZero(t, iptablesConnectionMark(0xa5c10000))
}

func TestIPTablesChainNamesFitKernelLimit(t *testing.T) {
	names := make([]string, 0, 4)
	egress, ingress, generationEgress, generationIngress := iptablesChainNames(
		net.ParseIP("255.255.255.255"), ^uint64(0),
	)
	names = append(names, egress, ingress, generationEgress, generationIngress)
	for _, name := range names {
		assert.LessOrEqual(t, len(name), 28, name)
	}
}

func TestIPTablesHooksCoverForwardedAndNodeLocalTraffic(t *testing.T) {
	hooks := aclIPTablesHooks(net.ParseIP("10.88.0.2"), "egress", "ingress")
	assert.Equal(t, []iptablesHook{
		{chain: "FORWARD", rule: []string{"-s", "10.88.0.2", "-j", "egress"}},
		{chain: "INPUT", rule: []string{"-s", "10.88.0.2", "-j", "egress"}},
		{chain: "FORWARD", rule: []string{"-d", "10.88.0.2", "-j", "ingress"}},
		{chain: "OUTPUT", rule: []string{"-d", "10.88.0.2", "-j", "ingress"}},
	}, hooks)
}
