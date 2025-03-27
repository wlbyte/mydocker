package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"log"

	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/cgroups"
	"github.com/wlbyte/mydocker/cgroups/subsystems"
	"github.com/wlbyte/mydocker/consts"
	"github.com/wlbyte/mydocker/container"
	"github.com/wlbyte/mydocker/utils"
)

var RunCommand = cli.Command{
	Name:  "run",
	Usage: "Create a container. eg: mydocker run -d|-it [-name containerName] [imageName] [command]",

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
		cli.StringFlag{
			Name:  "name",
			Usage: "container name, eg: run -name",
		},
	},
	Action: func(context *cli.Context) error {
		errFormat := "runCommand: %w"
		if len(context.Args()) < 2 {
			return fmt.Errorf(errFormat, errors.New("too few args"))
		}
		c := &container.Container{
			Name:     context.String("name"),
			TTY:      context.Bool("it"),
			Detach:   context.Bool("d"),
			Volume:   context.String("v"),
			CreateAt: time.Now().Format("2006-01-02 15:04:05"),
		}
		if c.TTY && c.Detach || (!c.TTY && !c.Detach) {
			return fmt.Errorf(errFormat, errors.New("choose flag between -it and -d"))
		}
		id, err := utils.HashStr(c)
		if err != nil {
			return fmt.Errorf(errFormat, err)
		}
		c.Id = id
		if c.Name == "" {
			if len(c.Id) > 12 {
				c.Name = c.Id[:12]
			} else {
				c.Name = c.Id
			}
		}
		c.ImageName = context.Args().Get(0)
		c.Cmds = context.Args().Tail()
		c.ResourceConfig = &subsystems.ResourceConfig{
			MemoryLimit: context.String("mem"),
			Cpus:        context.String("cpu"),
			CpuSet:      context.String("cpuset"),
		}
		run(c)
		return nil
	},
}

func run(c *container.Container) {
	parent, writePipe, err := container.NewParentProcess(c)
	if err != nil {
		log.Println("[error] run:", err)
		return
	}
	defer writePipe.Close()
	if err := parent.Start(); err != nil {
		log.Println("[error] run:", err)
		return
	}
	c.Pid = parent.Process.Pid
	c.Status = consts.STATUS_RUNNING
	if err := recordContainerInfo(c); err != nil {
		log.Printf("[error] run: %s", err)
		return
	}
	sendInitCommand(c.Cmds, writePipe)
	log.Println("[debug] send init command to pipe")
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	if err := cgroupManager.Set(c.ResourceConfig); err != nil {
		log.Println("[error] run:", err)
	}
	if err := cgroupManager.Apply(parent.Process.Pid, c.ResourceConfig); err != nil {
		log.Println("[error] run:", err)
	}
	if c.TTY {
		if err := parent.Wait(); err != nil {
			log.Printf("[error] run parent.Wait: %s", err)
		}
		log.Println("[debug] release resource")
		if err := cgroupManager.Destroy(); err != nil {
			log.Println("[error] run:", err)
		}
		log.Println("[debug] clear work dir")
		container.DelWorkspace(c)
		return
	}
	log.Println("[debug] run as a daemon")
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Printf("[debug] command: %s\n", command)
	if _, err := writePipe.WriteString(command); err != nil {
		log.Printf("[error] writePipe.WriteString: %s\n", err)
	}
	writePipe.Close()
}
