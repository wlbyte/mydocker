package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"log"

	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/cgroups"
	"github.com/wlbyte/mydocker/cgroups/subsystems"
	"github.com/wlbyte/mydocker/container"
	"github.com/wlbyte/mydocker/image"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit
			mydocker run -it [command]
			mydocker run -d -name [containerName] [imageName] [command]`,

	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "it",
			Usage: "enable tty, eg: run -it ",
		},
		cli.StringFlag{
			Name:  "mem",
			Usage: "memory limit, eg: run -mem 100m, {m|M|g|G}",
		},
		cli.StringFlag{
			Name:  "cpu",
			Usage: "cpu quota, eg: run -cpu 0.5", // 限制进程 cpu 使用率
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit,e.g.: run -cpuset 2,4", // 指定cpu位置
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "mount volume, eg: run -v containerDir:hostDir",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach, eg: run -d",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("runCommand: %w", errors.New("too few args"))
		}
		tty := context.Bool("it")
		detach := context.Bool("d")
		if tty && detach || (!tty && !detach) {
			return fmt.Errorf("runCommand: %w", errors.New("choose flag between -it and -d"))
		}
		var cmdSlice []string
		for _, arg := range context.Args() {
			cmdSlice = append(cmdSlice, arg)
		}
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("mem"),
			Cpus:        context.String("cpu"),
			CpuSet:      context.String("cpuset"),
		}
		volumePath := context.String("v")
		Run(tty, cmdSlice, resConf, volumePath)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		log.Println("[debug] init container")
		err := container.RunContainerInitProcess()
		if err != nil {
			log.Println("[error] initCommand:", err)
		}
		return err
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "build image",
	// Flags: []cli.Flag{
	// 	cli.StringFlag{
	// 		Name:  "name",
	// 		Usage: "build container to image. eg: commit -name 'mydocker'",
	// 	},
	// },
	Action: func(ctx *cli.Context) error {
		log.Println("[debug] build image")
		errFormat := "build image: %w"
		if len(ctx.Args()) < 1 {
			return fmt.Errorf(errFormat, errors.New("too few args"))
		}
		imageName := ctx.Args().Get(0)
		if err := image.BuildImage(imageName); err != nil {
			return fmt.Errorf(errFormat, err)
		}
		return nil
	},
}

func Run(tty bool, comArray []string, rs *subsystems.ResourceConfig, volumePath string) {
	parent, writePipe, err := container.NewParentProcess(tty, volumePath)
	if err != nil {
		log.Printf("[error] runCommand.Run: %v", err)
		return
	}
	defer writePipe.Close()
	if err := parent.Start(); err != nil {
		log.Printf("[error] parent.Start error: %s", err)
		return
	}
	sendInitCommand(comArray, writePipe)
	log.Println("[debug] send init command to pipe")
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	if err := cgroupManager.Set(rs); err != nil {
		log.Println("[debug] ", err)
	}
	if err := cgroupManager.Apply(parent.Process.Pid, rs); err != nil {
		log.Println("[debug] ", err)
	}
	if tty {
		if err := parent.Wait(); err != nil {
			log.Printf("[error] parent.Wait error: %s\n", err)
		}
		log.Println("[debug] release resource")
		if err := cgroupManager.Destroy(); err != nil {
			log.Println("[debug] ", err)
		}
		log.Println("[debug] clear work dir")
		container.DelWorkspace("/root", "/root/merged", volumePath)
	}
	log.Println("[debug] run as a daemon")
}

// sendInitCommand 通过writePipe将指令发送给子进程
func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Printf("[debug] command: %s\n", command)
	if _, err := writePipe.WriteString(command); err != nil {
		log.Printf("[error] writePipe.WriteString: %s\n", err)
	}
	writePipe.Close()
}
