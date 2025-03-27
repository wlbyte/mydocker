package cmd

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/consts"
	"github.com/wlbyte/mydocker/container"
)

var ListCommand = cli.Command{
	Name:  "ps",
	Usage: "list container info",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "a",
			Usage: "show all container, eg: ps -a ",
		},
	},
	Action: func(context *cli.Context) error {
		fmt.Println("[debug] list container info")
		all := context.Bool("a")
		fs := findJsonFilePathAll(consts.PATH_CONTAINER)
		cis := getContainerInfoAll(fs)
		printContainerInfo(cis, all)
		return nil
	},
}

func printContainerInfo(ci []*container.Container, all bool) {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, err := fmt.Fprint(w, "CONTAINER ID\tIMAGE\tCOMMAND\tCREATED\tSTATUS\tPID\tNAME\n")
	if err != nil {
		log.Println("[error] printContainerInfo:", err)
	}

	for _, c := range ci {
		if c.Status != consts.STATUS_RUNNING && !all {
			continue
		}
		printID := c.Id
		if len(c.Id) > 12 {
			printID = c.Id[:12]
		}
		_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
			printID,
			c.ImageName,
			c.Cmds,
			c.CreateAt,
			c.Status,
			c.Pid,
			c.Name,
		)
		if err != nil {
			log.Println("[error] printContainerInfo:", err)
		}
	}
	if err := w.Flush(); err != nil {
		log.Println("[error] printContainerInfo:", err)
	}
}
