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

package main

import (
	"fmt"

	"github.com/urfave/cli"
)

// HealthzCmd is a cli command to check the health of the sandbox service.
var HealthzCmd = cli.Command{
	Name:  "check",
	Usage: "Print the health status of the sandbox service",
	Action: func(context *cli.Context) error {
		client, err := NewSandboxClient(context)
		if err != nil {
			return err
		}
		fmt.Printf("Healthz status: %+v \n", client.Healthz())
		return nil
	},
}

var healthServiceConfig = `{
	"loadBalancingPolicy": "round_robin",
	"healthCheckConfig": {
		"serviceName": "sandbox"
	}
}`
