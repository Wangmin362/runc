package libcontainer

import (
	"errors"
	"io"
	"math"
	"os"

	"github.com/opencontainers/runc/libcontainer/configs"
)

var errInvalidProcess = errors.New("invalid process")

// TODO 这个接口用来抽象什么的？
type processOperations interface {
	wait() (*os.ProcessState, error)
	signal(sig os.Signal) error
	pid() int
}

// Process specifies the configuration and IO for a process inside
// a container.
// TODO 用于抽象用户的进程， libcontainer是如何抽象linux的进程的？赋予了什么能力？
type Process struct {
	// The command to be run followed by any arguments.
	// 启动参数
	Args []string

	// Env specifies the environment variables for the process.
	// TODO 为此进程设置的环境变量  这个环境变量包含那些环境变量？是config.json文件中指定的环境变量？如果上层应用，譬如K8S像容器注入了
	//环境变量，此时这些环境变量会放到这里么？
	Env []string

	// User will set the uid and gid of the executing process running inside the container
	// local to the container's user and group configuration.
	// 启动这个进程对应的用户，进程本质上也是操作系统抽象出来的一种资源，既然是资源，那么必定是属于某个用户的
	User string

	// AdditionalGroups specifies the gids that should be added to supplementary groups
	// in addition to those that the user belongs to.
	// TODO 这玩意应该是linux的概念
	AdditionalGroups []string

	// Cwd will change the processes current working directory inside the container's rootfs.
	// current work directory，也就是工作目录，再Dockerfile中通过Workdir指令可以指定这个参数
	Cwd string

	// Stdin is a pointer to a reader which provides the standard input stream.
	// TODO 标准输入
	Stdin io.Reader

	// Stdout is a pointer to a writer which receives the standard output stream.
	// TODO 标准输出
	Stdout io.Writer

	// Stderr is a pointer to a writer which receives the standard error stream.
	// TODO 标准错误
	Stderr io.Writer

	// ExtraFiles specifies additional open files to be inherited by the container
	// TODO 这玩意被抽象出来干嘛的？
	ExtraFiles []*os.File

	// Initial sizings for the console
	ConsoleWidth  uint16
	ConsoleHeight uint16

	// Capabilities specify the capabilities to keep when executing the process inside the container
	// All capabilities not specified will be dropped from the processes capability mask
	// Linux的能力，可以用于限制进程
	Capabilities *configs.Capabilities

	// AppArmorProfile specifies the profile to apply to the process and is
	// changed at the time the process is execed
	// 也是用于限制进程的技术，和SELinux是一个类型的技术
	AppArmorProfile string

	// Label specifies the label to apply to the process.  It is commonly used by selinux
	// 用于SELinux
	Label string

	// NoNewPrivileges controls whether processes can gain additional privileges.
	// TODO 用于控制Linux是否能够获取超级权限
	NoNewPrivileges *bool

	// Rlimits specifies the resource limits, such as max open files, to set in the container
	// If Rlimits are not set, the container will inherit rlimits from the parent process
	// 依然是用于限制进程的能力
	Rlimits []configs.Rlimit

	// ConsoleSocket provides the masterfd console.
	ConsoleSocket *os.File

	// Init specifies whether the process is the first process in the container.
	// 用于设置当前进程是否是容器中的第一个进程
	Init bool

	ops processOperations

	// 日志级别
	LogLevel string

	// SubCgroupPaths specifies sub-cgroups to run the process in.
	// Map keys are controller names, map values are paths (relative to
	// container's top-level cgroup).
	//
	// If empty, the default top-level container's cgroup is used.
	//
	// For cgroup v2, the only key allowed is "".
	// cGroup控制
	SubCgroupPaths map[string]string
}

// Wait waits for the process to exit.
// Wait releases any resources associated with the Process
func (p Process) Wait() (*os.ProcessState, error) {
	if p.ops == nil {
		return nil, errInvalidProcess
	}
	return p.ops.wait()
}

// Pid returns the process ID
func (p Process) Pid() (int, error) {
	// math.MinInt32 is returned here, because it's invalid value
	// for the kill() system call.
	if p.ops == nil {
		return math.MinInt32, errInvalidProcess
	}
	return p.ops.pid(), nil
}

// Signal sends a signal to the Process.
func (p Process) Signal(sig os.Signal) error {
	if p.ops == nil {
		return errInvalidProcess
	}
	return p.ops.signal(sig)
}

// IO holds the process's STDIO
type IO struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser
}
