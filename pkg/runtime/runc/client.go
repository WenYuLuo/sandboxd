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

package runc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/inclusionAI/sandboxd/internal/util"
)

// commandFailure classifies only diagnostics from a completed runc process.
// Launch, context and filesystem errors must never look like container absence.
type commandFailure struct {
	cause   error
	detail  string
	missing bool
	stopped bool
}

func (e *commandFailure) Error() string { return e.cause.Error() + ": " + e.detail }
func (e *commandFailure) Unwrap() error { return e.cause }

// State is the stable subset of `runc list --format json` used by sandboxd.
type State struct {
	ID      string `json:"id"`
	PID     int    `json:"pid"`
	Status  string `json:"status"`
	Bundle  string `json:"bundle"`
	Created string `json:"created"`
}

// Client executes one runc binary against a private state root.
type Client struct {
	binary string
	root   string
}

func NewClient(binary, root string) *Client {
	return &Client{binary: binary, root: root}
}

func (c *Client) command(ctx context.Context, args ...string) ([]byte, error) {
	all := append([]string{"--root", c.root}, args...)
	command := exec.CommandContext(ctx, c.binary, all...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err == nil {
		return stdout.Bytes(), nil
	}
	details := make([]string, 0, 2)
	if output := strings.TrimSpace(stdout.String()); output != "" {
		details = append(details, output)
	}
	if output := strings.TrimSpace(stderr.String()); output != "" {
		details = append(details, output)
	}
	if ctx.Err() != nil {
		return stdout.Bytes(), errors.Join(ctx.Err(), err)
	}
	if len(details) == 0 {
		return stdout.Bytes(), err
	}
	detail := strings.Join(details, "\n")
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return stdout.Bytes(), fmt.Errorf("%w: %s", err, detail)
	}
	message := strings.ToLower(detail)
	missing := strings.Contains(message, "container does not exist")
	stopped := missing || strings.Contains(message, "container is not running") ||
		strings.Contains(message, "container is stopped") || strings.Contains(message, "container not running")
	return stdout.Bytes(), &commandFailure{cause: err, detail: detail, missing: missing, stopped: stopped}
}

func (c *Client) List(ctx context.Context) ([]State, error) {
	output, err := c.command(ctx, "list", "--format", "json")
	if err != nil {
		return nil, err
	}
	var states []State
	if err := json.Unmarshal(output, &states); err != nil {
		return nil, fmt.Errorf("decode runc list: %w", err)
	}
	return states, nil
}

func (c *Client) State(ctx context.Context, id string) (State, error) {
	output, err := c.command(ctx, "state", id)
	if err != nil {
		return State{}, err
	}
	var state State
	if err := json.Unmarshal(output, &state); err != nil {
		return State{}, fmt.Errorf("decode runc state: %w", err)
	}
	return state, nil
}

func (c *Client) Kill(ctx context.Context, id, signal string, all bool) error {
	args := []string{"kill"}
	if all {
		args = append(args, "--all")
	}
	args = append(args, id, signal)
	_, err := c.command(ctx, args...)
	return err
}

func (c *Client) Delete(ctx context.Context, id string, force bool) error {
	if id == "" || filepath.Base(id) != id {
		return fmt.Errorf("invalid runc state id %q", id)
	}
	statePath, err := util.JoinWithinRoot(c.root, id)
	if err != nil {
		return err
	}
	args := []string{"delete"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, id)
	_, err = c.command(ctx, args...)
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || !IsNotFound(err) {
		return err
	}
	// A failed runc run can create its directory before writing state.json.
	// The CLI reports that container as absent without removing the directory.
	// Only remove an empty directory after confirmed absence; preserve state
	// files, symlinks and unexpected entries for recovery instead of masking them.
	info, statErr := os.Lstat(statePath)
	if os.IsNotExist(statErr) {
		return nil
	}
	if statErr != nil {
		return fmt.Errorf("inspect runc state directory: %w", statErr)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("runc state path is not a directory: %s", statePath)
	}
	if removeErr := os.Remove(statePath); removeErr != nil && !os.IsNotExist(removeErr) {
		return fmt.Errorf("remove empty runc state directory: %w", removeErr)
	}
	return nil
}

// IsNotFound reports the stable class of errors produced for missing state.
func IsNotFound(err error) bool {
	var failure *commandFailure
	return errors.As(err, &failure) && failure.missing
}

// IsNotRunning reports a confirmed stopped or missing container.
func IsNotRunning(err error) bool {
	var failure *commandFailure
	return errors.As(err, &failure) && failure.stopped
}

// IgnoreMissing turns idempotent absence into success.
func IgnoreMissing(err error) error {
	if IsNotFound(err) {
		return nil
	}
	return err
}
