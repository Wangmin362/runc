package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	exactArgs = iota // 精确参数校验
	minArgs          // 最小参数校验
	maxArgs          // 最大参数校验
)

// 用于校验参数的数量，检验方式可以是精确参数校验、最小参数个数的校验、最大参数个数的校验
func checkArgs(context *cli.Context, expected, checkType int) error {
	var err error
	cmdName := context.Command.Name
	switch checkType {
	case exactArgs:
		if context.NArg() != expected {
			err = fmt.Errorf("%s: %q requires exactly %d argument(s)", os.Args[0], cmdName, expected)
		}
	case minArgs:
		if context.NArg() < expected {
			err = fmt.Errorf("%s: %q requires a minimum of %d argument(s)", os.Args[0], cmdName, expected)
		}
	case maxArgs:
		if context.NArg() > expected {
			err = fmt.Errorf("%s: %q requires a maximum of %d argument(s)", os.Args[0], cmdName, expected)
		}
	}

	if err != nil {
		fmt.Printf("Incorrect Usage.\n\n")
		_ = cli.ShowCommandHelp(context, cmdName)
		return err
	}
	return nil
}

func logrusToStderr() bool {
	l, ok := logrus.StandardLogger().Out.(*os.File)
	return ok && l.Fd() == os.Stderr.Fd()
}

// fatal prints the error's details if it is a libcontainer specific error type
// then exits the program with an exit status of 1.
func fatal(err error) {
	fatalWithCode(err, 1)
}

func fatalWithCode(err error, ret int) {
	// Make sure the error is written to the logger.
	logrus.Error(err)
	if !logrusToStderr() {
		fmt.Fprintln(os.Stderr, err)
	}

	os.Exit(ret)
}

// setupSpec performs initial setup based on the cli.Context for the container
// 加载指定bundle目录中的config.json配置文件，如果用户没有设置bundle
func setupSpec(context *cli.Context) (*specs.Spec, error) {
	// 获取用户指定的bundle参数
	bundle := context.String("bundle")
	if bundle != "" { // 如果用户设置了bundle，说明bundle目录和当前的工作目录不一样，因此需要切换工作目录
		// 切换当前的工作目录为指定的bundle目录，类似于cd <bundle>这个动作，其实估计也是通过系统调用切换工作目录
		if err := os.Chdir(bundle); err != nil {
			return nil, err
		}
	}

	// 如果用户没有指定bundle目录，那么默认就是当前的工作目录就是bundle目录，如果当前目录并不是标准的bundle，那么肯定会有问题

	// 一个标准的bundle目录应该包含两个部分，其中一部分是config.json文件，另外一部分是根文件系统，这里其实就是在加载当前bundle目录中的config.json文件
	spec, err := loadSpec(specConfig)
	if err != nil {
		return nil, err
	}
	return spec, nil
}

// revisePidFile 修复用户传递的pidFile参数，如果用户没有设置这个参数，直接返回；如果设置了，并且路径用的相对路径，那么转换为绝对路径
func revisePidFile(context *cli.Context) error {
	// 获取用户指定的参数pid-file
	pidFile := context.String("pid-file")
	// 如果没有指定，直接返回
	if pidFile == "" {
		return nil
	}

	// convert pid-file to an absolute path so we can write to the right
	// file after chdir to bundle
	// 如果指定了pid-file，那么设置为绝对路径之后返回
	pidFile, err := filepath.Abs(pidFile)
	if err != nil {
		return err
	}
	return context.Set("pid-file", pidFile)
}

// reviseRootDir convert the root to absolute path
func reviseRootDir(context *cli.Context) error {
	root := context.GlobalString("root")
	if root == "" {
		return nil
	}

	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	return context.GlobalSet("root", root)
}

// parseBoolOrAuto returns (nil, nil) if s is empty or "auto"
func parseBoolOrAuto(s string) (*bool, error) {
	if s == "" || strings.ToLower(s) == "auto" {
		return nil, nil
	}
	b, err := strconv.ParseBool(s)
	return &b, err
}
