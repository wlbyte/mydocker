package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/image"
)

var CommitCommand = cli.Command{
	Name:  "commit",
	Usage: "mydocker commit containerID imageName",
	Action: func(ctx *cli.Context) error {
		log.Println("[debug] build image")
		errFormat := "build image: %w"
		if len(ctx.Args()) < 2 {
			return fmt.Errorf(errFormat, errors.New("too few args"))
		}
		containerID := ctx.Args().Get(0)
		imageName := ctx.Args().Get(1)
		c := GetContainerInfo(containerID)
		if c == nil {
			return fmt.Errorf(errFormat, errors.New("container not exist"))
		}
		if err := image.BuildImage(c.Id, imageName); err != nil {
			return fmt.Errorf(errFormat, err)
		}
		return nil
	},
}
