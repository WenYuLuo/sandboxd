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
	"flag"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func TestAddressFlagHasNoEnvironmentAlias(t *testing.T) {
	app := newApp()
	for _, candidate := range app.Flags {
		address, ok := candidate.(cli.StringFlag)
		if ok && address.Name == "address, a" {
			if address.EnvVar != "" {
				t.Fatalf("address EnvVar = %q, want empty", address.EnvVar)
			}
			return
		}
	}
	t.Fatal("address flag not found")
}

func TestDebugFlagControlsLogLevel(t *testing.T) {
	originalLevel := logrus.GetLevel()
	t.Cleanup(func() { logrus.SetLevel(originalLevel) })

	app := newApp()
	flags := flag.NewFlagSet("sbox-test", flag.ContinueOnError)
	flags.Bool("debug", false, "")
	ctx := cli.NewContext(app, flags, nil)
	if err := app.Before(ctx); err != nil {
		t.Fatal(err)
	}
	if got := logrus.GetLevel(); got != logrus.InfoLevel {
		t.Fatalf("default log level = %s, want info", got)
	}

	if err := flags.Set("debug", "true"); err != nil {
		t.Fatal(err)
	}
	if err := app.Before(ctx); err != nil {
		t.Fatal(err)
	}
	if got := logrus.GetLevel(); got != logrus.DebugLevel {
		t.Fatalf("debug log level = %s, want debug", got)
	}
}
