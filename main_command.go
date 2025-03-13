package main

import (
	"fmt"
	"os"
	"strings"

	"log"

	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/cgroups"
	"github.com/wlbyte/mydocker/cgroups/subsystems"
	"github.com/wlbyte/mydocker/container"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit
			mydocker run -it [command]
			mydocker run -d -name [containerName] [imageName] [command]`,

	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "it",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name:  "mem",
			Usage: "memory limit, eg: -mem 100m, {m|M|g|G}",
		},
		cli.StringFlag{
			Name:  "cpu",
			Usage: "cpu quota, eg: -cpu 0.5", // 限制进程 cpu 使用率
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit,e.g.: -cpuset 2,4", // 指定cpu位置
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}
		var cmdSlice []string
		for _, arg := range context.Args() {
			cmdSlice = append(cmdSlice, arg)
		}
		tty := context.Bool("it")
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("mem"),
			Cpus:        context.String("cpu"),
			CpuSet:      context.String("cpuset"),
		}
		Run(tty, cmdSlice, resConf)
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
			log.Println("[error] ", err)
		}
		return err
	},
}

func Run(tty bool, comArray []string, rs *subsystems.ResourceConfig) {
	parent, writePipe, err := container.NewParentProcess(tty)
	if err != nil {
		log.Printf("[error] New parent error: %v\n", err)
		return
	}
	defer writePipe.Close()
	if err := parent.Start(); err != nil {
		log.Printf("[error] parent.Start error: %s\n", err)
		return
	}
	sendInitCommand(comArray, writePipe)
	log.Println("[debug] send init command to pipe")
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer func() {
		log.Println("[debug] release resource")
		if err := cgroupManager.Destroy(); err != nil {
			log.Println("[debug] ", err)
		}
		log.Println("[debug] clear work dir")
		container.DelWorkspace("/root", "/root/merged")
	}()
	if err := cgroupManager.Set(rs); err != nil {
		log.Println("[debug] ", err)
	}
	if err := cgroupManager.Apply(parent.Process.Pid, rs); err != nil {
		log.Println("[debug] ", err)
	}
	if err := parent.Wait(); err != nil {
		log.Printf("[error] parent.Wait error: %s\n", err)
	}
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
