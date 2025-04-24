package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/consts"
	"github.com/wlbyte/mydocker/container"
)

var RemoveCommand = cli.Command{
	Name:  "rm",
	Usage: "remove container stopped",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "f",
			Usage: "force remove container, eg: rm -f ID ",
		},
	},
	Action: func(ctx *cli.Context) error {
		log.Println("[debug] remove container")
		errFormat := "rmCommand: %w"
		if len(ctx.Args()) < 1 {
			return fmt.Errorf(errFormat, errors.New("too few args"))
		}
		f := ctx.Bool("f")
		if err := rmContainer(ctx.Args(), f); err != nil {
			return fmt.Errorf(errFormat, err)
		}
		return nil
	},
}

func rmContainer(containerIDs []string, force bool) error {
	errFormat := "rmContainer: %w"
	for _, id := range containerIDs {
		c := GetContainerInfo(id)
		if c == nil {
			return fmt.Errorf(errFormat, errors.New("conainter is not exist"))
		}
		if c.Status == consts.STATUS_RUNNING {
			if !force {
				return fmt.Errorf(errFormat, errors.New("container must be stopped"))
			}
			if err := stopContainer(id); err != nil {
				return fmt.Errorf(errFormat, err)
			}
		}
		container.DelWorkspace(c)

		if err := os.Remove(findJsonFilePath(c.Id, consts.PATH_NETWORK_ENDPOINT)); err != nil {
			return fmt.Errorf(errFormat, err)
		}
	}

	return nil
}
