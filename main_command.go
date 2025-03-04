package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/container"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit
			mydocker run -it [command]
			mydocker run -d -name [containerName] [imageName] [command]`,

	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "it", // 简单起见，这里把 -i 和 -t 参数合并成一个
			Usage: "enable tty",
		},
		// cli.StringFlag{
		// 	Name:  "mem", // 限制进程内存使用量，为了避免和 stress 命令的 -m 参数冲突 这里使用 -mem,到时候可以看下解决冲突的方法
		// 	Usage: "memory limit,e.g.: -mem 100m",
		// },
		// cli.StringFlag{
		// 	Name:  "cpu",
		// 	Usage: "cpu quota,e.g.: -cpu 100", // 限制进程 cpu 使用率
		// },
		// cli.StringFlag{
		// 	Name:  "cpuset",
		// 	Usage: "cpuset limit,e.g.: -cpuset 2,4", // 指定cpu位置
		// },
		// cli.StringFlag{ // 数据卷
		// 	Name:  "v",
		// 	Usage: "volume,e.g.: -v /ect/conf:/etc/conf",
		// },
		// cli.BoolFlag{
		// 	Name:  "d",
		// 	Usage: "detach container,run background",
		// },
		// // 提供run后面的-name指定容器名字参数
		// cli.StringFlag{
		// 	Name:  "name",
		// 	Usage: "container name，e.g.: -name mycontainer",
		// },
		// cli.StringSliceFlag{
		// 	Name:  "e",
		// 	Usage: "set environment,e.g. -e name=mydocker",
		// },
		// cli.StringFlag{
		// 	Name:  "net",
		// 	Usage: "container network,e.g. -net testbr",
		// },
		// cli.StringSliceFlag{
		// 	Name:  "p",
		// 	Usage: "port mapping,e.g. -p 8080:80 -p 30336:3306",
		// },
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}
		cmd := context.Args().Get(0)
		tty := context.Bool("it")
		Run(tty, cmd)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		log.Infof("init")
		cmd := context.Args().Get(0)
		log.Infof("command: %s", cmd)
		err := container.RunContainerInitProcess(cmd, nil)
		return err
	},
}
