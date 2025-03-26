package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/wlbyte/mydocker/constants"
	_ "github.com/wlbyte/mydocker/nsenter"
)

const (
	EnvExecPid = "mydocker_pid"
	EnvExecCmd = "mydocker_cmd"
)

func execContainer(containerId string, comdArray []string) {
	f := findJsonFilePath(containerId, constants.CONTAINER_PATH)
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
