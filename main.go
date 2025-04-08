package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/cmd"
	"github.com/wlbyte/mydocker/consts"
)

const usage = `mydocker is a simple container runtime implementation.
			   The purpose of this project is to learn how docker works
			   and how to write a docker by ourselves Enjoy it, just for fun.`

func main() {
	if err := initDir(); err != nil {
		log.Fatal("[error] mydocker: ", err)
	}
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
		cmd.NetworkCommand,
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

func initDir() error {
	errFormat := "initDir %s: %w"
	// 创建数据目录
	if err := os.MkdirAll(consts.PATH_CONTAINER, consts.MODE_0755); err != nil {
		return fmt.Errorf(errFormat, consts.PATH_CONTAINER, err)
	}
	if err := os.MkdirAll(consts.PATH_FS_ROOT, consts.MODE_0755); err != nil {
		return fmt.Errorf(errFormat, consts.PATH_FS_ROOT, err)
	}
	if err := os.MkdirAll(consts.PATH_IMAGE, consts.MODE_0755); err != nil {
		return fmt.Errorf(errFormat, consts.PATH_IMAGE, err)
	}
	if err := os.MkdirAll(consts.PATH_IPAM, consts.MODE_0755); err != nil {
		return fmt.Errorf(errFormat, consts.PATH_IPAM, err)
	}
	if err := os.MkdirAll(consts.PATH_NETWORK_NETWORK, consts.MODE_0755); err != nil {
		return fmt.Errorf(errFormat, consts.PATH_NETWORK_NETWORK, err)
	}
	return nil
}
