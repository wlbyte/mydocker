package cmd

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/consts"
	"github.com/wlbyte/mydocker/network"
	"golang.org/x/sys/unix"
)

var StopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop container",
	Action: func(ctx *cli.Context) error {
		log.Println("[debug] stop container")
		errFormat := "stopCommand: %w"
		if len(ctx.Args()) < 1 {
			return fmt.Errorf(errFormat, errors.New("too few args"))
		}
		containerID := ctx.Args().Get(0)
		if err := stopContainer(containerID); err != nil {
			return fmt.Errorf(errFormat, err)
		}
		return nil
	},
}

func stopContainer(containerID string) error {
	errFormat := "stopContainer: %w"
	c := GetContainerInfo(containerID)
	if c == nil {
		return fmt.Errorf(errFormat, errors.New("conainter is not exist"))
	}
	e := GetEndpointInfo(c.Id)
	if err := network.DelConnect(c, e); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := unix.Kill(c.Pid, unix.SIGTERM); err != nil {
		if !strings.Contains(err.Error(), "no such process") {
			return fmt.Errorf(errFormat, err)
		}
	}
	c.Pid = 0
	c.Status = consts.STATUS_STOPPED
	if err := recordContainerInfo(c); err != nil {
		return fmt.Errorf(errFormat, err)
	}

	return nil
}
