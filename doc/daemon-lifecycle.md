# Daemon lifecycle and failed Start

A sandboxd daemon restart preserves externally managed sandboxes. Shutdown stops readiness and sandbox monitors but does not call Delete for existing sandboxes. Their filesystem mounts, cgroups and network objects remain available for the next daemon to recover from persistent records. When there are no sandboxes, Shutdown releases unused infrastructure. Explicit Delete remains responsible for terminating a sandbox and reclaiming its resources. This is distinct from stopping an orchestrator that intentionally deletes its workloads first.

A failed Start may include the gRPC trailer `sandboxd-start-settled: true`. The marker is set after the Start implementation and its deferred rollback have returned. It proves that this invocation cannot subsequently create a new backend; it does not prove that rollback succeeded or that no backend remains. Callers must inspect and clean up any remaining backend before releasing admission resources. A transport interruption without the marker remains an unknown outcome; a momentarily empty List is insufficient proof of settlement.

An empty object prefix is valid for an object at the bucket root. Explicit endpoint and bucket values therefore override the object storage template even when ObjectPrefix is empty. Object storage signature compatibility is separate from local HTTP HEAD/Range verification.

## Kata PTY mounts

Before rootfs preparation, the Kata handler adds a `devpts` mount at `/dev/pts` to the serialized OCI bundle when none exists. The mount uses a private `newinstance` with `ptmxmode=0666`, so the guest's `/dev/ptmx` link can resolve to `/dev/pts/ptmx`. An explicitly configured devpts mount retains its options; a conflicting mount type is rejected before starting the shim.

Tests verify the persisted spec, idempotence, custom options and rejection. A native guest `pty.openpty()` and SDK PTY test are separate runtime acceptance requirements; serialized-spec tests alone do not prove a Kata guest can boot.
