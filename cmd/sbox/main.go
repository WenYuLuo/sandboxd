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
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := newApp()
	if err := app.Run(os.Args); err != nil {
		if err.Error() != "" {
			fmt.Fprintf(os.Stderr, "sbox: %s\n", err)
		}
		os.Exit(commandExitCode(err))
	}
}

func commandExitCode(err error) int {
	var exitCoder cli.ExitCoder
	if errors.As(err, &exitCoder) {
		return exitCoder.ExitCode()
	}
	return 1
}
