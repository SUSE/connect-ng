package collectors

// Most of the code has been ported from
// https://github.com/GoogleContainerTools/kaniko/blob/ad1896a680f9bf26a4388ba0e702e7aac91fddc2/pkg/util/proc/proc.go,
// which at the same time was adapted and expanded from
// https://github.com/genuinetools/bpfd/blob/a4bfa5e3e9d1bfdbc56268a36a0714911ae9b6bf/proc/proc.go.

import (
	"os"
	"regexp"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
)

// ContainerRuntime is the type for the various container runtime strings.
type Runtime string

const (
	// RuntimeDocker is the string for the docker runtime.
	RuntimeDocker Runtime = "docker"
	// RuntimeRkt is the string for the rkt runtime.
	RuntimeRkt Runtime = "rkt"
	// RuntimeNspawn is the string for the systemd-nspawn runtime.
	RuntimeNspawn Runtime = "systemd-nspawn"
	// RuntimeLXC is the string for the lxc runtime.
	RuntimeLXC Runtime = "lxc"
	// RuntimeLXCLibvirt is the string for the lxc-libvirt runtime.
	RuntimeLXCLibvirt Runtime = "lxc-libvirt"
	// RuntimeOpenVZ is the string for the openvz runtime.
	RuntimeOpenVZ Runtime = "openvz"
	// RuntimeKubernetes is the string for the kubernetes runtime.
	RuntimeKubernetes Runtime = "kube"
	// RuntimeGarden is the string for the garden runtime.
	RuntimeGarden Runtime = "garden"
	// RuntimePodman is the string for the podman runtime.
	RuntimePodman Runtime = "podman"
	// RuntimeGVisor is the string for the gVisor (runsc) runtime.
	RuntimeGVisor Runtime = "gvisor"
	// RuntimeFirejail is the string for the firejail runtime.
	RuntimeFirejail Runtime = "firejail"
	// RuntimeWSL is the string for the Windows Subsystem for Linux runtime.
	RuntimeWSL Runtime = "wsl"
	// RuntimeNotFound is the string for when no container runtime is found.
	RuntimeNotFound Runtime = "not-found"
)

var (
	// ContainerRuntimes contains all the container runtimes.
	containerRuntimes = []Runtime{
		RuntimeDocker,
		RuntimeRkt,
		RuntimeNspawn,
		RuntimeLXC,
		RuntimeLXCLibvirt,
		RuntimeOpenVZ,
		RuntimeKubernetes,
		RuntimeGarden,
		RuntimePodman,
		RuntimeGVisor,
		RuntimeFirejail,
		RuntimeWSL,
	}
)

// GetContainerRuntime returns the container runtime the process is running in.
func GetContainerRuntime() Runtime {
	// read the cgroups file
	a := util.ReadFileString("/proc/self/cgroup")
	runtime := getContainerRuntime(a)
	if runtime != RuntimeNotFound {
		return runtime
	}

	// /proc/vz exists in container and outside of the container, /proc/bc only outside of the container.
	if util.FileExists("/proc/vz") && !util.FileExists("/proc/bc") {
		return RuntimeOpenVZ
	}

	// /__runsc_containers__ directory is present in gVisor containers.
	if util.FileExists("/__runsc_containers__") {
		return RuntimeGVisor
	}

	// firejail runs with `firejail` as pid 1.
	// As firejail binary cannot be run with argv[0] != "firejail"
	// it's okay to rely on cmdline.
	a = util.ReadFileString("/proc/1/cmdline")
	runtime = getContainerRuntime(a)
	if runtime != RuntimeNotFound {
		return runtime
	}

	// WSL has /proc/version_signature starting with "Microsoft".
	a = util.ReadFileString("/proc/version_signature")
	if strings.HasPrefix(a, "Microsoft") {
		return RuntimeWSL
	}

	a = os.Getenv("container")
	runtime = getContainerRuntime(a)
	if runtime != RuntimeNotFound {
		return runtime
	}

	// PID 1 might have dropped this information into a file in /run.
	// Read from /run/systemd/container since it is better than accessing /proc/1/environ,
	// which needs CAP_SYS_PTRACE
	a = util.ReadFileString("/run/systemd/container")
	runtime = getContainerRuntime(a)
	if runtime != RuntimeNotFound {
		return runtime
	}

	// Check for container specific files
	runtime = detectContainerFiles()
	if runtime != RuntimeNotFound {
		return runtime
	}

	// Docker was not detected at this point.
	// An overlay mount on "/" may indicate we're under containerd or other runtime.
	a = util.ReadFileString("/proc/mounts")
	if m, _ := regexp.MatchString("^[^ ]+ / overlay", a); m {
		return RuntimeKubernetes
	}

	return RuntimeNotFound
}

// Related implementation: https://github.com/systemd/systemd/blob/6604fb0207ee10e8dc05d67f6fe45de0b193b5c4/src/basic/virt.c#L523-L549
func detectContainerFiles() Runtime {
	files := []struct {
		runtime  Runtime
		location string
	}{
		// https://github.com/containers/podman/issues/6192
		// https://github.com/containers/podman/issues/3586#issuecomment-661918679
		{RuntimePodman, "/run/.containerenv"},
		// https://github.com/moby/moby/issues/18355
		{RuntimeDocker, "/.dockerenv"},
		// Detect the presence of a serviceaccount secret mounted in the default location
		{RuntimeKubernetes, "/var/run/secrets/kubernetes.io/serviceaccount"},
	}

	for i := range files {
		if util.FileExists(files[i].location) {
			return files[i].runtime
		}
	}

	return RuntimeNotFound
}

func getContainerRuntime(input string) Runtime {
	if len(strings.TrimSpace(input)) < 1 {
		return RuntimeNotFound
	}

	for _, runtime := range containerRuntimes {
		if strings.Contains(input, string(runtime)) {
			return runtime
		}
	}

	return RuntimeNotFound
}

// NOTE: here starts the actual code from the collector which have not been
// ported from the aforementioned packages.

type ContainerRuntime struct{}

func (cr ContainerRuntime) run(arch string) (Result, error) {
	if runtime := GetContainerRuntime(); runtime != RuntimeNotFound {
		return Result{"container_runtime": runtime}, nil
	}
	return NoResult, nil
}
