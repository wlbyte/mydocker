package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/consts"
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
		log.Println("[debug] exec container command")
		if os.Getenv(EnvExecPid) != "" {
			log.Printf("[debug] pid callback pid %d\n", os.Getgid())
			return nil
		}
		if len(context.Args()) < 2 {
			return fmt.Errorf("missing container id or command")
		}
		cId := context.Args().Get(0)
		var cmds []string
		cmds = append(cmds, context.Args().Tail()...)
		execContainer(cId, cmds)
		return nil
	},
}

func execContainer(containerId string, comdArray []string) {
	f := findJsonFilePath(containerId, consts.PATH_CONTAINER)
	c := getContainerInfo(f)
	if c == nil {
		log.Println("[error] execContainer getContainerInfo: container info is nil")
		os.Exit(1)
	}
	pid := c.Pid
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmdStr := strings.Join(comdArray, " ")
	log.Printf("[debug] container pid: %d command: %s\n", pid, cmdStr)
	os.Setenv(EnvExecPid, strconv.Itoa(pid))
	os.Setenv(EnvExecCmd, cmdStr)
	if err := cmd.Run(); err != nil {
		log.Println("[error] exec container:", err)
	}
}
