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
	"fmt"
	"net"
	"strconv"

	"github.com/coreos/go-iptables/iptables"
	"github.com/inclusionAI/sandboxd/config"
	"github.com/inclusionAI/sandboxd/pkg/networkmanager"
)

type BridgeNetworkManager struct{}

type iptablesClient interface {
	Append(table, chain string, rulespec ...string) error
	AppendUnique(table, chain string, rulespec ...string) error
	Delete(table, chain string, rulespec ...string) error
	DeleteIfExists(table, chain string, rulespec ...string) error
	Exists(table, chain string, rulespec ...string) (bool, error)
}

var newIPTablesClient = func() (iptablesClient, error) {
	return iptables.New()
}

// SetupSNATRules implements networkmanager.NetworkManager.
func (BridgeNetworkManager) SetupSNATRules(ipRange string) error {
	// add follow iptable rule: iptables -t nat -A POSTROUTING -s 172.17.0.0/16 -j MASQUERADE
	ipt, err := newIPTablesClient()
	if err != nil {
		return err
	}
	// check if rule exists.
	if exists, err := ipt.Exists("nat", "POSTROUTING", "-s", ipRange, "-j", "MASQUERADE"); err != nil {
		return err
	} else if exists {
		return nil
	}

	// create rule.
	return ipt.Append("nat", "POSTROUTING", "-s", ipRange, "-j", "MASQUERADE")
}

// CleanupSNATRules implements networkmanager.NetworkManager.
func (BridgeNetworkManager) CleanupSNATRules(ipRange string) error {
	// clean iptable rule if exists.
	ipt, err := newIPTablesClient()
	if err != nil {
		return err
	}
	// check if rule exists.
	if exists, err := ipt.Exists("nat", "POSTROUTING", "-s", ipRange, "-j", "MASQUERADE"); err != nil {
		return err
	} else if !exists {
		return nil
	}

	// delete rule.
	return ipt.Delete("nat", "POSTROUTING", "-s", ipRange, "-j", "MASQUERADE")
}

// SetupNetworkRulesForActivating implements networkmanager.NetworkManager.
func (BridgeNetworkManager) SetupNetworkRulesForActivating(ip net.IP, envId string) error {
	return nil
}

// CleanupNetworkRulesForActivating implements networkmanager.NetworkManager.
func (BridgeNetworkManager) CleanupNetworkRulesForActivating(ip net.IP) error {
	return nil
}

// SetupDNATRule implements networkmanager.NetworkManager.
func (BridgeNetworkManager) SetupDNATRule(protocol string, dstPort uint16, targetIP string, targetPort uint16) error {
	ipt, err := newIPTablesClient()
	if err != nil {
		return err
	}

	dstPortStr := strconv.FormatUint(uint64(dstPort), 10)
	targetPortStr := strconv.FormatUint(uint64(targetPort), 10)
	toDest := fmt.Sprintf("%s:%s", targetIP, targetPortStr)

	preroutingRule := []string{
		"-p", protocol,
		"--dport", dstPortStr,
		"-j", "DNAT",
		"--to-destination", toDest,
	}
	forwardRule := []string{
		"-p", protocol,
		"-d", targetIP,
		"--dport", targetPortStr,
		"-j", "ACCEPT",
	}

	// PREROUTING handles traffic entering the node network namespace.
	if err := ipt.AppendUnique("nat", "PREROUTING", preroutingRule...); err != nil {
		return fmt.Errorf("failed to add PREROUTING DNAT rule: %v", err)
	}

	if err := ipt.AppendUnique("filter", "FORWARD", forwardRule...); err != nil {
		_ = ipt.Delete("nat", "PREROUTING", preroutingRule...)
		return fmt.Errorf("failed to add FORWARD rule: %v", err)
	}

	return nil
}

// SetupLocalDNATRule forwards callers that share sandboxd's network namespace.
func (BridgeNetworkManager) SetupLocalDNATRule(
	protocol string,
	dstPort uint16,
	targetIP string,
	targetPort uint16,
) error {
	ipt, err := newIPTablesClient()
	if err != nil {
		return err
	}
	dstPortStr := strconv.FormatUint(uint64(dstPort), 10)
	toDest := fmt.Sprintf("%s:%d", targetIP, targetPort)
	outputRule := []string{
		"-p", protocol,
		"-m", "addrtype",
		"--dst-type", "LOCAL",
		"--dport", dstPortStr,
		"-j", "DNAT",
		"--to-destination", toDest,
	}
	if err := ipt.AppendUnique("nat", "OUTPUT", outputRule...); err != nil {
		return fmt.Errorf("failed to add OUTPUT DNAT rule: %v", err)
	}
	return nil
}

// CleanupDNATRule implements networkmanager.NetworkManager.
func (BridgeNetworkManager) CleanupDNATRule(protocol string, dstPort uint16, targetIP string, targetPort uint16) error {
	ipt, err := newIPTablesClient()
	if err != nil {
		return err
	}

	dstPortStr := strconv.FormatUint(uint64(dstPort), 10)
	targetPortStr := strconv.FormatUint(uint64(targetPort), 10)
	toDest := fmt.Sprintf("%s:%s", targetIP, targetPortStr)

	preroutingRule := []string{
		"-p", protocol,
		"--dport", dstPortStr,
		"-j", "DNAT",
		"--to-destination", toDest,
	}
	forwardRule := []string{
		"-p", protocol,
		"-d", targetIP,
		"--dport", targetPortStr,
		"-j", "ACCEPT",
	}

	// Best-effort: remove both ingress rules and report the first error.
	var firstErr error

	if err := ipt.DeleteIfExists("nat", "PREROUTING", preroutingRule...); err != nil {
		firstErr = fmt.Errorf("failed to delete PREROUTING DNAT rule: %v", err)
	}

	if err := ipt.DeleteIfExists("filter", "FORWARD", forwardRule...); err != nil && firstErr == nil {
		firstErr = fmt.Errorf("failed to delete FORWARD rule: %v", err)
	}

	return firstErr
}

// CleanupLocalDNATRule removes a local-caller rule whether or not the feature
// remains enabled, so configuration changes do not strand compatible rules.
func (BridgeNetworkManager) CleanupLocalDNATRule(
	protocol string,
	dstPort uint16,
	targetIP string,
	targetPort uint16,
) error {
	ipt, err := newIPTablesClient()
	if err != nil {
		return err
	}
	dstPortStr := strconv.FormatUint(uint64(dstPort), 10)
	toDest := fmt.Sprintf("%s:%d", targetIP, targetPort)
	outputRule := []string{
		"-p", protocol,
		"-m", "addrtype",
		"--dst-type", "LOCAL",
		"--dport", dstPortStr,
		"-j", "DNAT",
		"--to-destination", toDest,
	}
	if err := ipt.DeleteIfExists("nat", "OUTPUT", outputRule...); err != nil {
		return fmt.Errorf("failed to delete OUTPUT DNAT rule: %v", err)
	}
	return nil
}

func init() {
	networkmanager.Register(config.NatBackendIptables, &BridgeNetworkManager{})
}
