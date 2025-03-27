package cmd

import (
	"fmt"
	"log"

	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/container"
)

var InitCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		log.Println("[debug] init container")
		err := container.RunContainerInitProcess()
		if err != nil {
			return fmt.Errorf("initCommand: %w", err)
		}
		return nil
	},
}
