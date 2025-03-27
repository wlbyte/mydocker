package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/consts"
)

var LogsCommand = cli.Command{
	Name:  "logs",
	Usage: "get container logs",
	Action: func(context *cli.Context) error {
		log.Println("[debug] get container logs")
		if len(context.Args()) < 1 {
			return fmt.Errorf("logsCommand: %w", errors.New("no container ID"))
		}
		cSubID := context.Args().Get(0)
		f := findJsonFilePath(cSubID, consts.PATH_CONTAINER)
		c := getContainerInfo(f)
		if c != nil {
			logFile := fmt.Sprintf("%s/%s/%s.log", consts.PATH_CONTAINER, c.Id, c.Id)
			bs, err := os.ReadFile(logFile)
			if err != nil {
				return fmt.Errorf("logsCommand: %w", err)
			}
			fmt.Printf("%s\n", bs)
		}
		return nil
	},
}
