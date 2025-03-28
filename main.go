package main

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/cmd"
)

const usage = `mydocker is a simple container runtime implementation.
			   The purpose of this project is to learn how docker works
			   and how to write a docker by ourselves Enjoy it, just for fun.`

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = usage

	app.Commands = []cli.Command{
		cmd.InitCommand,
		cmd.RunCommand,
		cmd.CommitCommand,
		cmd.ListCommand,
		cmd.LogsCommand,
		cmd.ExecCommand,
		cmd.StopCommand,
		cmd.RemoveCommand,
	}

	app.Before = func(context *cli.Context) error {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal("[error] mydocker: ", err)
	}
}
