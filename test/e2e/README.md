# sandboxd E2E

This flow validates the public runsc adapter without an AKernel node image.

It performs:

1. `go test ./...`
2. builds `output/sandboxd` and `output/sbox`
3. copies an externally installed runsc to `output/runsc`
4. builds a minimal image containing sandboxd, sbox, runsc, iproute2,
   iptables, and busybox
5. runs a privileged container and verifies start, list, inspect, exec, bind
   mounts, sandbox networking, stats, and delete

Use the tested `runsc release-20260706.0`. The adapter reads runsc state and
uses gVisor control RPCs, so another release is not assumed compatible.

```bash
RUNSC_BINARY=/usr/local/bin/runsc bash test/e2e/run.sh

# Equivalent Make target.
RUNSC_BINARY=/usr/local/bin/runsc make e2e
```

Requirements:

- an accessible Docker daemon
- a Linux host with cgroup v1
- permission to run privileged containers with the host cgroup namespace
- a usable iptables nat table

Set `RUN_UNIT_TESTS=0` to skip the unit-test step when rerunning only the
privileged scenario.
