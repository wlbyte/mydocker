package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/consts"
	"github.com/wlbyte/mydocker/container"
	_ "github.com/wlbyte/mydocker/nsenter"
)

const (
	EnvExecPid = "mydocker_pid"
	EnvExecCmd = "mydocker_cmd"
)

var ExecCommand = cli.Command{
	Name:  "exec",
	Usage: "exec container command",
	Action: func(context *cli.Context) error {
		errFormat := "execCommand: %w"
		if os.Getenv(EnvExecPid) != "" {
			return nil
		}
		if len(context.Args()) < 2 {
			return fmt.Errorf(errFormat, errors.New("missing containerID or command"))
		}
		cId := context.Args().Get(0)
		var cmds []string
		cmds = append(cmds, context.Args().Tail()...)
		if err := execContainer(cId, cmds); err != nil {
			return fmt.Errorf(errFormat, err)
		}
		return nil
	},
}

func execContainer(containerId string, comdArray []string) error {
	errFormat := "execContainer: %w"
	f := findJsonFilePath(containerId, consts.PATH_CONTAINER)
	c := getContainerInfo(f)
	if c == nil {
		return fmt.Errorf(errFormat, container.ErrContainerNotExist)
	}
	pid := c.Pid
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmdStr := strings.Join(comdArray, " ")
	os.Setenv(EnvExecPid, strconv.Itoa(pid))
	os.Setenv(EnvExecCmd, cmdStr)
	envs, err := getEnvsById(containerId)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	cmd.Env = append(os.Environ(), envs...)
	log.Printf("[debug] container pid: %d, command: %s\n", pid, cmdStr)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func getEnvsById(containerID string) ([]string, error) {
	errFormat := "getEnvsByID: %w"
	c := GetContainerInfo(containerID)
	if c == nil {
		return nil, fmt.Errorf(errFormat, container.ErrContainerNotExist)
	}
	bs, err := os.ReadFile("/proc/" + strconv.Itoa(c.Pid) + "/environ")
	if err != nil {
		return nil, fmt.Errorf(errFormat, err)
	}
	return strings.Split(string(bs), "\u0000"), nil
}
