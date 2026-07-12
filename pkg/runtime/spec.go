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

// Portions of this file are derived from the Open Container Initiative
// runtime-spec project and are licensed under the Apache License 2.0.

package runtime

import "os"

// Spec is the simple-version base only for linux configuration for the container.
type Spec struct {
	// Version of the Open Container Initiative Runtime Specification with which the bundle complies.
	Version string `json:"ociVersion"`
	// Process configures the container process.
	Process *Process `json:"process,omitempty"`
	// Root configures the container's root filesystem.
	Root *Root `json:"root,omitempty"`
	// Hostname configures the container's hostname.
	Hostname string `json:"hostname,omitempty"`
	// Mounts configures additional mounts (on top of Root).
	Mounts []Mount `json:"mounts,omitempty"`
	// Hooks configures callbacks for container lifecycle events.
	Hooks *Hooks `json:"hooks,omitempty" platform:"linux,solaris"`
	// Annotations contains arbitrary metadata for the container.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Linux is platform-specific configuration for Linux based containers.
	Linux *Linux `json:"linux,omitempty" platform:"linux"`
}

func (s *Spec) DeepCopy() *Spec {
	if s == nil {
		return nil
	}
	out := new(Spec)
	out.Version = s.Version
	if s.Process != nil {
		out.Process = s.Process.DeepCopy()
	}
	if s.Root != nil {
		out.Root = s.Root.DeepCopy()
	}
	out.Hostname = s.Hostname
	if s.Mounts != nil {
		out.Mounts = make([]Mount, len(s.Mounts))
		for i := range s.Mounts {
			out.Mounts[i] = *s.Mounts[i].DeepCopy()
		}
	}
	if s.Hooks != nil {
		out.Hooks = s.Hooks.DeepCopy()
	}
	if s.Annotations != nil {
		out.Annotations = make(map[string]string, len(s.Annotations))
		for k, v := range s.Annotations {
			out.Annotations[k] = v
		}
	}
	if s.Linux != nil {
		out.Linux = s.Linux.DeepCopy()
	}
	return out
}

// Process contains information to start a specific application inside the container.
type Process struct {
	// Terminal creates an interactive terminal for the container.
	Terminal bool `json:"terminal,omitempty"`
	// ConsoleSize specifies the size of the console.
	ConsoleSize *Box `json:"consoleSize,omitempty"`
	// User specifies user information for the process.
	User User `json:"user"`
	// Args specifies the binary and arguments for the application to execute.
	Args []string `json:"args,omitempty"`
	// CommandLine specifies the full command line for the application to execute on Windows.
	CommandLine string `json:"commandLine,omitempty" platform:"windows"`
	// Env populates the process environment for the process.
	Env []string `json:"env,omitempty"`
	// Cwd is the current working directory for the process and must be
	// relative to the container's root.
	Cwd string `json:"cwd"`
	// Capabilities are Linux capabilities that are kept for the process.
	Capabilities *LinuxCapabilities `json:"capabilities,omitempty" platform:"linux"`
	// Rlimits specifies rlimit options to apply to the process.
	Rlimits []POSIXRlimit `json:"rlimits,omitempty" platform:"linux,solaris"`
	// NoNewPrivileges controls whether additional privileges could be gained by processes in the container.
	NoNewPrivileges bool `json:"noNewPrivileges,omitempty" platform:"linux"`
	// ApparmorProfile specifies the apparmor profile for the container.
	ApparmorProfile string `json:"apparmorProfile,omitempty" platform:"linux"`
	// Specify an oom_score_adj for the container.
	OOMScoreAdj *int `json:"oomScoreAdj,omitempty" platform:"linux"`
	// SelinuxLabel specifies the selinux context that the container process is run as.
	SelinuxLabel string `json:"selinuxLabel,omitempty" platform:"linux"`
}

// DeepCopy returns a deep copy of the Process structure.
func (p *Process) DeepCopy() *Process {
	if p == nil {
		return nil
	}
	out := new(Process)
	out.Terminal = p.Terminal
	if p.ConsoleSize != nil {
		out.ConsoleSize = p.ConsoleSize.DeepCopy()
	}
	out.User = p.User
	if p.Args != nil {
		out.Args = make([]string, len(p.Args))
		copy(out.Args, p.Args)
	}
	out.CommandLine = p.CommandLine
	if p.Env != nil {
		out.Env = make([]string, len(p.Env))
		copy(out.Env, p.Env)
	}
	out.Cwd = p.Cwd
	if p.Capabilities != nil {
		out.Capabilities = p.Capabilities.DeepCopy()
	}
	if p.Rlimits != nil {
		out.Rlimits = make([]POSIXRlimit, len(p.Rlimits))
		copy(out.Rlimits, p.Rlimits)
	}
	out.NoNewPrivileges = p.NoNewPrivileges
	out.ApparmorProfile = p.ApparmorProfile
	if p.OOMScoreAdj != nil {
		out.OOMScoreAdj = new(int)
		*out.OOMScoreAdj = *p.OOMScoreAdj
	}
	out.SelinuxLabel = p.SelinuxLabel
	return out
}

// LinuxCapabilities specifies the list of allowed capabilities that are kept for a process.
// http://man7.org/linux/man-pages/man7/capabilities.7.html
type LinuxCapabilities struct {
	// Bounding is the set of capabilities checked by the kernel.
	Bounding []string `json:"bounding,omitempty" platform:"linux"`
	// Effective is the set of capabilities checked by the kernel.
	Effective []string `json:"effective,omitempty" platform:"linux"`
	// Inheritable is the capabilities preserved across execve.
	Inheritable []string `json:"inheritable,omitempty" platform:"linux"`
	// Permitted is the limiting superset for effective capabilities.
	Permitted []string `json:"permitted,omitempty" platform:"linux"`
	// Ambient is the ambient set of capabilities that are kept.
	Ambient []string `json:"ambient,omitempty" platform:"linux"`
}

// DeepCopy returns a deep copy of the LinuxCapabilities structure.
func (l *LinuxCapabilities) DeepCopy() *LinuxCapabilities {
	if l == nil {
		return nil
	}
	out := new(LinuxCapabilities)
	if l.Bounding != nil {
		out.Bounding = make([]string, len(l.Bounding))
		copy(out.Bounding, l.Bounding)
	}
	if l.Effective != nil {
		out.Effective = make([]string, len(l.Effective))
		copy(out.Effective, l.Effective)
	}
	if l.Inheritable != nil {
		out.Inheritable = make([]string, len(l.Inheritable))
		copy(out.Inheritable, l.Inheritable)
	}
	if l.Permitted != nil {
		out.Permitted = make([]string, len(l.Permitted))
		copy(out.Permitted, l.Permitted)
	}
	if l.Ambient != nil {
		out.Ambient = make([]string, len(l.Ambient))
		copy(out.Ambient, l.Ambient)
	}
	return out
}

// Box specifies dimensions of a rectangle. Used for specifying the size of a console.
type Box struct {
	// Height is the vertical dimension of a box.
	Height uint `json:"height"`
	// Width is the horizontal dimension of a box.
	Width uint `json:"width"`
}

// DeepCopy returns a deep copy of the Box structure.
func (b *Box) DeepCopy() *Box {
	if b == nil {
		return nil
	}
	out := new(Box)
	out.Height = b.Height
	out.Width = b.Width
	return out
}

// User specifies specific user (and group) information for the container process.
type User struct {
	// UID is the user id.
	UID uint32 `json:"uid" platform:"linux,solaris"`
	// GID is the group id.
	GID uint32 `json:"gid" platform:"linux,solaris"`
	// Umask is the umask for the init process.
	Umask *uint32 `json:"umask,omitempty" platform:"linux,solaris"`
	// AdditionalGids are additional group ids set for the container's process.
	AdditionalGids []uint32 `json:"additionalGids,omitempty" platform:"linux,solaris"`
	// Username is the user name.
	Username string `json:"username,omitempty" platform:"windows"`
}

// Root contains information about the container's root filesystem on the host.
type Root struct {
	// Path is the absolute path to the container's root filesystem.
	Path string `json:"path"`
	// Readonly makes the root filesystem for the container readonly before the process is executed.
	Readonly bool `json:"readonly,omitempty"`
}

// DeepCopy returns a deep copy of the Root structure.
func (r *Root) DeepCopy() *Root {
	if r == nil {
		return nil
	}
	out := new(Root)
	out.Path = r.Path
	out.Readonly = r.Readonly
	return out
}

// Mount specifies a mount for a container.
type Mount struct {
	// Destination is the absolute path where the mount will be placed in the container.
	Destination string `json:"destination"`
	// Type specifies the mount kind.
	Type string `json:"type,omitempty" platform:"linux,solaris"`
	// Source specifies the source path of the mount.
	Source string `json:"source,omitempty"`
	// Options are fstab style mount options.
	Options []string `json:"options,omitempty"`
}

// DeepCopy returns a deep copy of the Mount structure.
func (m *Mount) DeepCopy() *Mount {
	if m == nil {
		return nil
	}
	out := new(Mount)
	out.Destination = m.Destination
	out.Type = m.Type
	out.Source = m.Source
	if m.Options != nil {
		out.Options = make([]string, len(m.Options))
		copy(out.Options, m.Options)
	}
	return out
}

// Hook specifies a command that is run at a particular event in the lifecycle of a container
type Hook struct {
	Path    string   `json:"path"`
	Args    []string `json:"args,omitempty"`
	Env     []string `json:"env,omitempty"`
	Timeout *int     `json:"timeout,omitempty"`
}

// DeepCopy returns a deep copy of the Hook structure.
func (h *Hook) DeepCopy() *Hook {
	if h == nil {
		return nil
	}
	out := new(Hook)
	out.Path = h.Path
	if h.Args != nil {
		out.Args = make([]string, len(h.Args))
		copy(out.Args, h.Args)
	}
	if h.Env != nil {
		out.Env = make([]string, len(h.Env))
		copy(out.Env, h.Env)
	}
	if h.Timeout != nil {
		out.Timeout = new(int)
		*out.Timeout = *h.Timeout
	}
	return out
}

// Hooks specifies a command that is run in the container at a particular event in the lifecycle of a container
// Hooks for container setup and teardown
type Hooks struct {
	// Prestart is Deprecated. Prestart is a list of hooks to be run before the container process is executed.
	// It is called in the Runtime Namespace
	Prestart []Hook `json:"prestart,omitempty"`
	// CreateRuntime is a list of hooks to be run after the container has been created but before pivot_root or any equivalent operation has been called
	// It is called in the Runtime Namespace
	CreateRuntime []Hook `json:"createRuntime,omitempty"`
	// CreateContainer is a list of hooks to be run after the container has been created but before pivot_root or any equivalent operation has been called
	// It is called in the Container Namespace
	CreateContainer []Hook `json:"createContainer,omitempty"`
	// StartContainer is a list of hooks to be run after the start operation is called but before the container process is started
	// It is called in the Container Namespace
	StartContainer []Hook `json:"startContainer,omitempty"`
	// Poststart is a list of hooks to be run after the container process is started.
	// It is called in the Runtime Namespace
	Poststart []Hook `json:"poststart,omitempty"`
	// Poststop is a list of hooks to be run after the container process exits.
	// It is called in the Runtime Namespace
	Poststop []Hook `json:"poststop,omitempty"`
}

// DeepCopy returns a deep copy of the Hooks structure.
func (h *Hooks) DeepCopy() *Hooks {
	if h == nil {
		return nil
	}
	out := new(Hooks)
	if h.Prestart != nil {
		out.Prestart = make([]Hook, len(h.Prestart))
		for i := range h.Prestart {
			out.Prestart[i] = *h.Prestart[i].DeepCopy()
		}
	}
	if h.CreateRuntime != nil {
		out.CreateRuntime = make([]Hook, len(h.CreateRuntime))
		for i := range h.CreateRuntime {
			out.CreateRuntime[i] = *h.CreateRuntime[i].DeepCopy()
		}
	}
	if h.CreateContainer != nil {
		out.CreateContainer = make([]Hook, len(h.CreateContainer))
		for i := range h.CreateContainer {
			out.CreateContainer[i] = *h.CreateContainer[i].DeepCopy()
		}
	}
	if h.StartContainer != nil {
		out.StartContainer = make([]Hook, len(h.StartContainer))
		for i := range h.StartContainer {
			out.StartContainer[i] = *h.StartContainer[i].DeepCopy()
		}
	}
	if h.Poststart != nil {
		out.Poststart = make([]Hook, len(h.Poststart))
		for i := range h.Poststart {
			out.Poststart[i] = *h.Poststart[i].DeepCopy()
		}
	}
	if h.Poststop != nil {
		out.Poststop = make([]Hook, len(h.Poststop))
		for i := range h.Poststop {
			out.Poststop[i] = *h.Poststop[i].DeepCopy()
		}
	}
	return out
}

// Linux contains platform-specific configuration for Linux based containers.
type Linux struct {
	// UIDMapping specifies user mappings for supporting user namespaces.
	UIDMappings []LinuxIDMapping `json:"uidMappings,omitempty"`
	// GIDMapping specifies group mappings for supporting user namespaces.
	GIDMappings []LinuxIDMapping `json:"gidMappings,omitempty"`
	// Sysctl are a set of key value pairs that are set for the container on start
	Sysctl map[string]string `json:"sysctl,omitempty"`
	// Resources contain cgroup information for handling resource constraints
	// for the container
	Resources *LinuxResources `json:"resources,omitempty"`
	// CgroupsPath specifies the path to cgroups that are created and/or joined by the container.
	// The path is expected to be relative to the cgroups mountpoint.
	// If resources are specified, the cgroups at CgroupsPath will be updated based on resources.
	CgroupsPath string `json:"cgroupsPath,omitempty"`
	// Namespaces contains the namespaces that are created and/or joined by the container
	Namespaces []LinuxNamespace `json:"namespaces,omitempty"`
	// Devices are a list of device nodes that are created for the container
	Devices []LinuxDevice `json:"devices,omitempty"`
	// Seccomp specifies the seccomp security settings for the container.
	Seccomp *LinuxSeccomp `json:"seccomp,omitempty"`
	// RootfsPropagation is the rootfs mount propagation mode for the container.
	RootfsPropagation string `json:"rootfsPropagation,omitempty"`
	// MaskedPaths masks over the provided paths inside the container.
	MaskedPaths []string `json:"maskedPaths,omitempty"`
	// ReadonlyPaths sets the provided paths as RO inside the container.
	ReadonlyPaths []string `json:"readonlyPaths,omitempty"`
	// MountLabel specifies the selinux context for the mounts in the container.
	MountLabel string `json:"mountLabel,omitempty"`
	// IntelRdt contains Intel Resource Director Technology (RDT) information for
	// handling resource constraints (e.g., L3 cache, memory bandwidth) for the container
	IntelRdt *LinuxIntelRdt `json:"intelRdt,omitempty"`
	// Personality contains configuration for the Linux personality syscall
	Personality *LinuxPersonality `json:"personality,omitempty"`
}

// DeepCopy returns a deep copy of the Linux structure.
func (l *Linux) DeepCopy() *Linux {
	if l == nil {
		return nil
	}
	out := new(Linux)
	if l.UIDMappings != nil {
		out.UIDMappings = make([]LinuxIDMapping, len(l.UIDMappings))
		for i := range l.UIDMappings {
			out.UIDMappings[i] = *l.UIDMappings[i].DeepCopy()
		}
	}
	if l.GIDMappings != nil {
		out.GIDMappings = make([]LinuxIDMapping, len(l.GIDMappings))
		for i := range l.GIDMappings {
			out.GIDMappings[i] = *l.GIDMappings[i].DeepCopy()
		}
	}
	if l.Sysctl != nil {
		out.Sysctl = make(map[string]string, len(l.Sysctl))
		for k, v := range l.Sysctl {
			out.Sysctl[k] = v
		}
	}
	if l.Resources != nil {
		out.Resources = l.Resources.DeepCopy()
	}
	out.CgroupsPath = l.CgroupsPath
	if l.Namespaces != nil {
		out.Namespaces = make([]LinuxNamespace, len(l.Namespaces))
		for i := range l.Namespaces {
			out.Namespaces[i] = *l.Namespaces[i].DeepCopy()
		}
	}
	if l.Devices != nil {
		out.Devices = make([]LinuxDevice, len(l.Devices))
		for i := range l.Devices {
			out.Devices[i] = *l.Devices[i].DeepCopy()
		}
	}
	if l.Seccomp != nil {
		out.Seccomp = l.Seccomp.DeepCopy()
	}
	out.RootfsPropagation = l.RootfsPropagation
	if l.MaskedPaths != nil {
		out.MaskedPaths = make([]string, len(l.MaskedPaths))
		for i := range l.MaskedPaths {
			out.MaskedPaths[i] = l.MaskedPaths[i]
		}
	}
	if l.ReadonlyPaths != nil {
		out.ReadonlyPaths = make([]string, len(l.ReadonlyPaths))
		for i := range l.ReadonlyPaths {
			out.ReadonlyPaths[i] = l.ReadonlyPaths[i]
		}
	}
	out.MountLabel = l.MountLabel
	if l.IntelRdt != nil {
		out.IntelRdt = l.IntelRdt.DeepCopy()
	}
	if l.Personality != nil {
		out.Personality = l.Personality.DeepCopy()
	}
	return out
}

// LinuxNamespace is the configuration for a Linux namespace
type LinuxNamespace struct {
	// Type is the type of namespace
	Type LinuxNamespaceType `json:"type"`
	// Path is a path to an existing namespace persisted on disk that can be joined
	// and is of the same type
	Path string `json:"path,omitempty"`
}

func (n *LinuxNamespace) DeepCopy() *LinuxNamespace {
	if n == nil {
		return nil
	}
	out := new(LinuxNamespace)
	out.Type = n.Type
	out.Path = n.Path
	return out
}

// LinuxNamespaceType is one of the Linux namespaces
type LinuxNamespaceType string

const (
	// PIDNamespace for isolating process IDs
	PIDNamespace LinuxNamespaceType = "pid"
	// NetworkNamespace for isolating network devices, stacks, ports, etc
	NetworkNamespace LinuxNamespaceType = "network"
	// MountNamespace for isolating mount points
	MountNamespace LinuxNamespaceType = "mount"
	// IPCNamespace for isolating System V IPC, POSIX message queues
	IPCNamespace LinuxNamespaceType = "ipc"
	// UTSNamespace for isolating hostname and NIS domain name
	UTSNamespace LinuxNamespaceType = "uts"
	// UserNamespace for isolating user and group IDs
	UserNamespace LinuxNamespaceType = "user"
	// CgroupNamespace for isolating cgroup hierarchies
	CgroupNamespace LinuxNamespaceType = "cgroup"
)

// LinuxIDMapping specifies UID/GID mappings
type LinuxIDMapping struct {
	// ContainerID is the starting UID/GID in the container
	ContainerID uint32 `json:"containerID"`
	// HostID is the starting UID/GID on the host to be mapped to 'ContainerID'
	HostID uint32 `json:"hostID"`
	// Size is the number of IDs to be mapped
	Size uint32 `json:"size"`
}

func (m *LinuxIDMapping) DeepCopy() *LinuxIDMapping {
	if m == nil {
		return nil
	}
	out := new(LinuxIDMapping)
	out.ContainerID = m.ContainerID
	out.HostID = m.HostID
	out.Size = m.Size
	return out
}

// POSIXRlimit type and restrictions
type POSIXRlimit struct {
	// Type of the rlimit to set
	Type string `json:"type"`
	// Hard is the hard limit for the specified type
	Hard uint64 `json:"hard"`
	// Soft is the soft limit for the specified type
	Soft uint64 `json:"soft"`
}

// LinuxHugepageLimit structure corresponds to limiting kernel hugepages
type LinuxHugepageLimit struct {
	// Pagesize is the hugepage size
	// Format: "<size><unit-prefix>B' (e.g. 64KB, 2MB, 1GB, etc.)
	Pagesize string `json:"pageSize"`
	// Limit is the limit of "hugepagesize" hugetlb usage
	Limit uint64 `json:"limit"`
}

// LinuxInterfacePriority for network interfaces
type LinuxInterfacePriority struct {
	// Name is the name of the network interface
	Name string `json:"name"`
	// Priority for the interface
	Priority uint32 `json:"priority"`
}

// linuxBlockIODevice holds major:minor format supported in blkio cgroup
type linuxBlockIODevice struct {
	// Major is the device's major number.
	Major int64 `json:"major"`
	// Minor is the device's minor number.
	Minor int64 `json:"minor"`
}

// LinuxWeightDevice struct holds a `major:minor weight` pair for weightDevice
type LinuxWeightDevice struct {
	linuxBlockIODevice
	// Weight is the bandwidth rate for the device.
	Weight *uint16 `json:"weight,omitempty"`
	// LeafWeight is the bandwidth rate for the device while competing with the cgroup's child cgroups, CFQ scheduler only
	LeafWeight *uint16 `json:"leafWeight,omitempty"`
}

// LinuxThrottleDevice struct holds a `major:minor rate_per_second` pair
type LinuxThrottleDevice struct {
	linuxBlockIODevice
	// Rate is the IO rate limit per cgroup per device
	Rate uint64 `json:"rate"`
}

// LinuxBlockIO for Linux cgroup 'blkio' resource management
type LinuxBlockIO struct {
	// Specifies per cgroup weight
	Weight *uint16 `json:"weight,omitempty"`
	// Specifies tasks' weight in the given cgroup while competing with the cgroup's child cgroups, CFQ scheduler only
	LeafWeight *uint16 `json:"leafWeight,omitempty"`
	// Weight per cgroup per device, can override BlkioWeight
	WeightDevice []LinuxWeightDevice `json:"weightDevice,omitempty"`
	// IO read rate limit per cgroup per device, bytes per second
	ThrottleReadBpsDevice []LinuxThrottleDevice `json:"throttleReadBpsDevice,omitempty"`
	// IO write rate limit per cgroup per device, bytes per second
	ThrottleWriteBpsDevice []LinuxThrottleDevice `json:"throttleWriteBpsDevice,omitempty"`
	// IO read rate limit per cgroup per device, IO per second
	ThrottleReadIOPSDevice []LinuxThrottleDevice `json:"throttleReadIOPSDevice,omitempty"`
	// IO write rate limit per cgroup per device, IO per second
	ThrottleWriteIOPSDevice []LinuxThrottleDevice `json:"throttleWriteIOPSDevice,omitempty"`
}

func (i *LinuxBlockIO) DeepCopy() *LinuxBlockIO {
	if i == nil {
		return nil
	}
	out := new(LinuxBlockIO)
	out.Weight = i.Weight
	out.LeafWeight = i.LeafWeight
	if i.WeightDevice != nil {
		out.WeightDevice = make([]LinuxWeightDevice, len(i.WeightDevice))
		for j := range i.WeightDevice {
			out.WeightDevice[j] = i.WeightDevice[j]
		}
	}
	if i.ThrottleReadBpsDevice != nil {
		out.ThrottleReadBpsDevice = make([]LinuxThrottleDevice, len(i.ThrottleReadBpsDevice))
		for j := range i.ThrottleReadBpsDevice {
			out.ThrottleReadBpsDevice[j] = i.ThrottleReadBpsDevice[j]
		}
	}
	if i.ThrottleWriteBpsDevice != nil {
		out.ThrottleWriteBpsDevice = make([]LinuxThrottleDevice, len(i.ThrottleWriteBpsDevice))
		for j := range i.ThrottleWriteBpsDevice {
			out.ThrottleWriteBpsDevice[j] = i.ThrottleWriteBpsDevice[j]
		}
	}
	if i.ThrottleReadIOPSDevice != nil {
		out.ThrottleReadIOPSDevice = make([]LinuxThrottleDevice, len(i.ThrottleReadIOPSDevice))
		for j := range i.ThrottleReadIOPSDevice {
			out.ThrottleReadIOPSDevice[j] = i.ThrottleReadIOPSDevice[j]
		}
	}
	if i.ThrottleWriteIOPSDevice != nil {
		out.ThrottleWriteIOPSDevice = make([]LinuxThrottleDevice, len(i.ThrottleWriteIOPSDevice))
		for j := range i.ThrottleWriteIOPSDevice {
			out.ThrottleWriteIOPSDevice[j] = i.ThrottleWriteIOPSDevice[j]
		}
	}
	return out
}

// LinuxMemory for Linux cgroup 'memory' resource management
type LinuxMemory struct {
	// Memory limit (in bytes).
	Limit *int64 `json:"limit,omitempty"`
	// Memory reservation or soft_limit (in bytes).
	Reservation *int64 `json:"reservation,omitempty"`
	// Total memory limit (memory + swap).
	Swap *int64 `json:"swap,omitempty"`
	// Kernel memory limit (in bytes).
	Kernel *int64 `json:"kernel,omitempty"`
	// Kernel memory limit for tcp (in bytes)
	KernelTCP *int64 `json:"kernelTCP,omitempty"`
	// How aggressive the kernel will swap memory pages.
	Swappiness *uint64 `json:"swappiness,omitempty"`
	// DisableOOMKiller disables the OOM killer for out of memory conditions
	DisableOOMKiller *bool `json:"disableOOMKiller,omitempty"`
	// Enables hierarchical memory accounting
	UseHierarchy *bool `json:"useHierarchy,omitempty"`
}

func (m *LinuxMemory) DeepCopy() *LinuxMemory {
	if m == nil {
		return nil
	}
	out := new(LinuxMemory)
	out.Limit = m.Limit
	out.Reservation = m.Reservation
	out.Swap = m.Swap
	out.Kernel = m.Kernel
	out.KernelTCP = m.KernelTCP
	out.Swappiness = m.Swappiness
	out.DisableOOMKiller = m.DisableOOMKiller
	out.UseHierarchy = m.UseHierarchy
	return out
}

// LinuxCPU for Linux cgroup 'cpu' resource management
type LinuxCPU struct {
	// CPU shares (relative weight (ratio) vs. other cgroups with cpu shares).
	Shares *uint64 `json:"shares,omitempty"`
	// CPU hardcap limit (in usecs). Allowed cpu time in a given period.
	Quota *int64 `json:"quota,omitempty"`
	// CPU period to be used for hardcapping (in usecs).
	Period *uint64 `json:"period,omitempty"`
	// How much time realtime scheduling may use (in usecs).
	RealtimeRuntime *int64 `json:"realtimeRuntime,omitempty"`
	// CPU period to be used for realtime scheduling (in usecs).
	RealtimePeriod *uint64 `json:"realtimePeriod,omitempty"`
	// CPUs to use within the cpuset. Default is to use any CPU available.
	Cpus string `json:"cpus,omitempty"`
	// List of memory nodes in the cpuset. Default is to use any available memory node.
	Mems string `json:"mems,omitempty"`
}

func (c *LinuxCPU) DeepCopy() *LinuxCPU {
	if c == nil {
		return nil
	}
	out := new(LinuxCPU)
	out.Shares = c.Shares
	out.Quota = c.Quota
	out.Period = c.Period
	out.RealtimeRuntime = c.RealtimeRuntime
	out.RealtimePeriod = c.RealtimePeriod
	out.Cpus = c.Cpus
	out.Mems = c.Mems
	return out
}

// LinuxPids for Linux cgroup 'pids' resource management (Linux 4.3)
type LinuxPids struct {
	// Maximum number of PIDs. Default is "no limit".
	Limit int64 `json:"limit"`
}

func (p *LinuxPids) DeepCopy() *LinuxPids {
	if p == nil {
		return nil
	}
	out := new(LinuxPids)
	out.Limit = p.Limit
	return out
}

// LinuxNetwork identification and priority configuration
type LinuxNetwork struct {
	// Set class identifier for container's network packets
	ClassID *uint32 `json:"classID,omitempty"`
	// Set priority of network traffic for container
	Priorities []LinuxInterfacePriority `json:"priorities,omitempty"`
}

func (n *LinuxNetwork) DeepCopy() *LinuxNetwork {
	if n == nil {
		return nil
	}
	out := new(LinuxNetwork)
	out.ClassID = n.ClassID
	if n.Priorities != nil {
		out.Priorities = make([]LinuxInterfacePriority, len(n.Priorities))
		for i := range n.Priorities {
			out.Priorities[i] = n.Priorities[i]
		}
	}
	return out
}

// LinuxRdma for Linux cgroup 'rdma' resource management (Linux 4.11)
type LinuxRdma struct {
	// Maximum number of HCA handles that can be opened. Default is "no limit".
	HcaHandles *uint32 `json:"hcaHandles,omitempty"`
	// Maximum number of HCA objects that can be created. Default is "no limit".
	HcaObjects *uint32 `json:"hcaObjects,omitempty"`
}

// LinuxResources has container runtime resource constraints
type LinuxResources struct {
	// Devices configures the device allowlist.
	Devices []LinuxDeviceCgroup `json:"devices,omitempty"`
	// Memory restriction configuration
	Memory *LinuxMemory `json:"memory,omitempty"`
	// CPU resource restriction configuration
	CPU *LinuxCPU `json:"cpu,omitempty"`
	// Task resource restriction configuration.
	Pids *LinuxPids `json:"pids,omitempty"`
	// BlockIO restriction configuration
	BlockIO *LinuxBlockIO `json:"blockIO,omitempty"`
	// Hugetlb limit (in bytes)
	HugepageLimits []LinuxHugepageLimit `json:"hugepageLimits,omitempty"`
	// Network restriction configuration
	Network *LinuxNetwork `json:"network,omitempty"`
	// Rdma resource restriction configuration.
	// Limits are a set of key value pairs that define RDMA resource limits,
	// where the key is device name and value is resource limits.
	Rdma map[string]LinuxRdma `json:"rdma,omitempty"`
	// Unified resources.
	Unified map[string]string `json:"unified,omitempty"`
}

func (r *LinuxResources) DeepCopy() *LinuxResources {
	if r == nil {
		return nil
	}
	out := new(LinuxResources)
	out.Devices = make([]LinuxDeviceCgroup, len(r.Devices))
	for i := range r.Devices {
		out.Devices[i] = *r.Devices[i].DeepCopy()
	}
	if r.Memory != nil {
		out.Memory = r.Memory.DeepCopy()
	}
	if r.CPU != nil {
		out.CPU = r.CPU.DeepCopy()
	}
	if r.Pids != nil {
		out.Pids = r.Pids.DeepCopy()
	}
	if r.BlockIO != nil {
		out.BlockIO = r.BlockIO.DeepCopy()
	}
	out.HugepageLimits = make([]LinuxHugepageLimit, len(r.HugepageLimits))
	for i := range r.HugepageLimits {
		out.HugepageLimits[i] = r.HugepageLimits[i]
	}
	if r.Network != nil {
		out.Network = r.Network.DeepCopy()
	}
	out.Rdma = make(map[string]LinuxRdma, len(r.Rdma))
	for k, v := range r.Rdma {
		out.Rdma[k] = v
	}
	out.Unified = make(map[string]string, len(r.Unified))
	for k, v := range r.Unified {
		out.Unified[k] = v
	}
	return out
}

// LinuxDevice represents the mknod information for a Linux special device file
type LinuxDevice struct {
	// Path to the device.
	Path string `json:"path"`
	// Device type, block, char, etc.
	Type string `json:"type"`
	// Major is the device's major number.
	Major int64 `json:"major"`
	// Minor is the device's minor number.
	Minor int64 `json:"minor"`
	// FileMode permission bits for the device.
	FileMode *os.FileMode `json:"fileMode,omitempty"`
	// UID of the device.
	UID *uint32 `json:"uid,omitempty"`
	// Gid of the device.
	GID *uint32 `json:"gid,omitempty"`
}

func (d *LinuxDevice) DeepCopy() *LinuxDevice {
	if d == nil {
		return nil
	}
	out := new(LinuxDevice)
	out.Path = d.Path
	out.Type = d.Type
	out.Major = d.Major
	out.Minor = d.Minor
	if d.FileMode != nil {
		out.FileMode = new(os.FileMode)
		*out.FileMode = *d.FileMode
	}
	if d.UID != nil {
		out.UID = new(uint32)
		*out.UID = *d.UID
	}
	if d.GID != nil {
		out.GID = new(uint32)
		*out.GID = *d.GID
	}
	return out
}

// LinuxDeviceCgroup represents a device rule for the devices specified to
// the device controller
type LinuxDeviceCgroup struct {
	// Allow or deny
	Allow bool `json:"allow"`
	// Device type, block, char, etc.
	Type string `json:"type,omitempty"`
	// Major is the device's major number.
	Major *int64 `json:"major,omitempty"`
	// Minor is the device's minor number.
	Minor *int64 `json:"minor,omitempty"`
	// Cgroup access permissions format, rwm.
	Access string `json:"access,omitempty"`
}

func (c *LinuxDeviceCgroup) DeepCopy() *LinuxDeviceCgroup {
	if c == nil {
		return nil
	}
	out := new(LinuxDeviceCgroup)
	out.Allow = c.Allow
	out.Type = c.Type
	if c.Major != nil {
		out.Major = new(int64)
		*out.Major = *c.Major
	}
	if c.Minor != nil {
		out.Minor = new(int64)
		*out.Minor = *c.Minor
	}
	out.Access = c.Access
	return out
}

// LinuxPersonalityDomain refers to a personality domain.
type LinuxPersonalityDomain string

// LinuxPersonalityFlag refers to an additional personality flag. None are currently defined.
type LinuxPersonalityFlag string

// Define domain and flags for Personality
const (
	// PerLinux is the standard Linux personality
	PerLinux LinuxPersonalityDomain = "LINUX"
	// PerLinux32 sets personality to 32 bit
	PerLinux32 LinuxPersonalityDomain = "LINUX32"
)

// LinuxPersonality represents the Linux personality syscall input
type LinuxPersonality struct {
	// Domain for the personality
	Domain LinuxPersonalityDomain `json:"domain"`
	// Additional flags
	Flags []LinuxPersonalityFlag `json:"flags,omitempty"`
}

func (p *LinuxPersonality) DeepCopy() *LinuxPersonality {
	if p == nil {
		return nil
	}
	out := new(LinuxPersonality)
	out.Domain = p.Domain
	if p.Flags != nil {
		out.Flags = make([]LinuxPersonalityFlag, len(p.Flags))
		copy(out.Flags, p.Flags)
	}
	return out
}

// LinuxSeccomp represents syscall restrictions
type LinuxSeccomp struct {
	DefaultAction    LinuxSeccompAction `json:"defaultAction"`
	DefaultErrnoRet  *uint              `json:"defaultErrnoRet,omitempty"`
	Architectures    []Arch             `json:"architectures,omitempty"`
	Flags            []LinuxSeccompFlag `json:"flags,omitempty"`
	ListenerPath     string             `json:"listenerPath,omitempty"`
	ListenerMetadata string             `json:"listenerMetadata,omitempty"`
	Syscalls         []LinuxSyscall     `json:"syscalls,omitempty"`
}

func (s *LinuxSeccomp) DeepCopy() *LinuxSeccomp {
	if s == nil {
		return nil
	}
	out := new(LinuxSeccomp)
	out.DefaultAction = s.DefaultAction
	if s.DefaultErrnoRet != nil {
		out.DefaultErrnoRet = new(uint)
		*out.DefaultErrnoRet = *s.DefaultErrnoRet
	}
	if s.Architectures != nil {
		out.Architectures = make([]Arch, len(s.Architectures))
		copy(out.Architectures, s.Architectures)
	}
	if s.Flags != nil {
		out.Flags = make([]LinuxSeccompFlag, len(s.Flags))
		copy(out.Flags, s.Flags)
	}
	out.ListenerPath = s.ListenerPath
	out.ListenerMetadata = s.ListenerMetadata
	if s.Syscalls != nil {
		out.Syscalls = make([]LinuxSyscall, len(s.Syscalls))
		for i := range s.Syscalls {
			out.Syscalls[i] = *s.Syscalls[i].DeepCopy()
		}
	}
	return out
}

// Arch used for additional architectures
type Arch string

// LinuxSeccompFlag is a flag to pass to seccomp(2).
type LinuxSeccompFlag string

// Additional architectures permitted to be used for system calls
// By default only the native architecture of the kernel is permitted
const (
	ArchX86         Arch = "SCMP_ARCH_X86"
	ArchX86_64      Arch = "SCMP_ARCH_X86_64"
	ArchX32         Arch = "SCMP_ARCH_X32"
	ArchARM         Arch = "SCMP_ARCH_ARM"
	ArchAARCH64     Arch = "SCMP_ARCH_AARCH64"
	ArchMIPS        Arch = "SCMP_ARCH_MIPS"
	ArchMIPS64      Arch = "SCMP_ARCH_MIPS64"
	ArchMIPS64N32   Arch = "SCMP_ARCH_MIPS64N32"
	ArchMIPSEL      Arch = "SCMP_ARCH_MIPSEL"
	ArchMIPSEL64    Arch = "SCMP_ARCH_MIPSEL64"
	ArchMIPSEL64N32 Arch = "SCMP_ARCH_MIPSEL64N32"
	ArchPPC         Arch = "SCMP_ARCH_PPC"
	ArchPPC64       Arch = "SCMP_ARCH_PPC64"
	ArchPPC64LE     Arch = "SCMP_ARCH_PPC64LE"
	ArchS390        Arch = "SCMP_ARCH_S390"
	ArchS390X       Arch = "SCMP_ARCH_S390X"
	ArchPARISC      Arch = "SCMP_ARCH_PARISC"
	ArchPARISC64    Arch = "SCMP_ARCH_PARISC64"
	ArchRISCV64     Arch = "SCMP_ARCH_RISCV64"
)

// LinuxSeccompAction taken upon Seccomp rule match
type LinuxSeccompAction string

// Define actions for Seccomp rules
const (
	ActKill        LinuxSeccompAction = "SCMP_ACT_KILL"
	ActKillProcess LinuxSeccompAction = "SCMP_ACT_KILL_PROCESS"
	ActKillThread  LinuxSeccompAction = "SCMP_ACT_KILL_THREAD"
	ActTrap        LinuxSeccompAction = "SCMP_ACT_TRAP"
	ActErrno       LinuxSeccompAction = "SCMP_ACT_ERRNO"
	ActTrace       LinuxSeccompAction = "SCMP_ACT_TRACE"
	ActAllow       LinuxSeccompAction = "SCMP_ACT_ALLOW"
	ActLog         LinuxSeccompAction = "SCMP_ACT_LOG"
	ActNotify      LinuxSeccompAction = "SCMP_ACT_NOTIFY"
)

// LinuxSeccompOperator used to match syscall arguments in Seccomp
type LinuxSeccompOperator string

// Define operators for syscall arguments in Seccomp
const (
	OpNotEqual     LinuxSeccompOperator = "SCMP_CMP_NE"
	OpLessThan     LinuxSeccompOperator = "SCMP_CMP_LT"
	OpLessEqual    LinuxSeccompOperator = "SCMP_CMP_LE"
	OpEqualTo      LinuxSeccompOperator = "SCMP_CMP_EQ"
	OpGreaterEqual LinuxSeccompOperator = "SCMP_CMP_GE"
	OpGreaterThan  LinuxSeccompOperator = "SCMP_CMP_GT"
	OpMaskedEqual  LinuxSeccompOperator = "SCMP_CMP_MASKED_EQ"
)

// LinuxSeccompArg used for matching specific syscall arguments in Seccomp
type LinuxSeccompArg struct {
	Index    uint                 `json:"index"`
	Value    uint64               `json:"value"`
	ValueTwo uint64               `json:"valueTwo,omitempty"`
	Op       LinuxSeccompOperator `json:"op"`
}

// LinuxSyscall is used to match a syscall in Seccomp
type LinuxSyscall struct {
	Names    []string           `json:"names"`
	Action   LinuxSeccompAction `json:"action"`
	ErrnoRet *uint              `json:"errnoRet,omitempty"`
	Args     []LinuxSeccompArg  `json:"args,omitempty"`
}

func (s *LinuxSyscall) DeepCopy() *LinuxSyscall {
	if s == nil {
		return nil
	}
	out := new(LinuxSyscall)
	out.Names = make([]string, len(s.Names))
	copy(out.Names, s.Names)
	out.Action = s.Action
	if s.ErrnoRet != nil {
		out.ErrnoRet = new(uint)
		*out.ErrnoRet = *s.ErrnoRet
	}
	if s.Args != nil {
		out.Args = make([]LinuxSeccompArg, len(s.Args))
		for i := range s.Args {
			out.Args[i] = s.Args[i]
		}
	}
	return out
}

// LinuxIntelRdt has container runtime resource constraints for Intel RDT
// CAT and MBA features which introduced in Linux 4.10 and 4.12 kernel
type LinuxIntelRdt struct {
	// The identity for RDT Class of Service
	ClosID string `json:"closID,omitempty"`
	// The schema for L3 cache id and capacity bitmask (CBM)
	// Format: "L3:<cache_id0>=<cbm0>;<cache_id1>=<cbm1>;..."
	L3CacheSchema string `json:"l3CacheSchema,omitempty"`

	// The schema of memory bandwidth per L3 cache id
	// Format: "MB:<cache_id0>=bandwidth0;<cache_id1>=bandwidth1;..."
	// The unit of memory bandwidth is specified in "percentages" by
	// default, and in "MBps" if MBA Software Controller is enabled.
	MemBwSchema string `json:"memBwSchema,omitempty"`
}

func (r *LinuxIntelRdt) DeepCopy() *LinuxIntelRdt {
	if r == nil {
		return nil
	}
	out := new(LinuxIntelRdt)
	out.ClosID = r.ClosID
	out.L3CacheSchema = r.L3CacheSchema
	out.MemBwSchema = r.MemBwSchema
	return out
}
