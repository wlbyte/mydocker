package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"log"

	"github.com/urfave/cli"
	"github.com/wlbyte/mydocker/cgroups"
	"github.com/wlbyte/mydocker/cgroups/subsystems"
	"github.com/wlbyte/mydocker/constants"
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
		cli.StringFlag{
			Name:  "name",
			Usage: "container name, eg: run -name",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("runCommand: %w", errors.New("too few args"))
		}
		c := &container.Container{
			Id:       strconv.Itoa(rand.Int()),
			Name:     context.String("name"),
			TTY:      context.Bool("it"),
			Detach:   context.Bool("d"),
			Volume:   context.String("v"),
			CreateAt: time.Now().Format("2006-01-02 15:04:05"),
		}
		if c.TTY && c.Detach || (!c.TTY && !c.Detach) {
			return fmt.Errorf("runCommand: %w", errors.New("choose flag between -it and -d"))
		}
		if c.Name == "" {
			c.Name = c.Id
		}
		var cmds []string
		for _, arg := range context.Args() {
			cmds = append(cmds, arg)
		}
		c.Cmds = cmds
		c.ResourceConfig = &subsystems.ResourceConfig{
			MemoryLimit: context.String("mem"),
			Cpus:        context.String("cpu"),
			CpuSet:      context.String("cpuset"),
		}
		Run(c)
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

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list container info",
	Action: func(context *cli.Context) error {
		log.Println("[debug] list container info")
		fs := findAllJsonFile(constants.CONTAINER_PATH)
		cis := getContainerInfo(fs)
		printContainerInfo(cis)
		return nil
	},
}

func Run(c *container.Container) {
	parent, writePipe, err := container.NewParentProcess(c.TTY, c.Volume)
	if err != nil {
		log.Printf("[error] runCommand.Run: %v", err)
		return
	}
	defer writePipe.Close()
	if err := parent.Start(); err != nil {
		log.Printf("[error] parent.Start error: %s", err)
		return
	}
	c.Pid = parent.Process.Pid
	c.Status = "Running"
	if err := recordContainerInfo(c); err != nil {
		log.Printf("[error] run error: %s", err)
		return
	}
	sendInitCommand(c.Cmds, writePipe)
	log.Println("[debug] send init command to pipe")
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	if err := cgroupManager.Set(c.ResourceConfig); err != nil {
		log.Println("[debug] ", err)
	}
	if err := cgroupManager.Apply(parent.Process.Pid, c.ResourceConfig); err != nil {
		log.Println("[debug] ", err)
	}
	if c.TTY {
		if err := parent.Wait(); err != nil {
			log.Printf("[error] parent.Wait error: %s\n", err)
		}
		log.Println("[debug] release resource")
		if err := cgroupManager.Destroy(); err != nil {
			log.Println("[debug] ", err)
		}
		log.Println("[debug] clear work dir")
		container.DelWorkspace("/root", "/root/merged", c.Volume, c.Id)
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

func recordContainerInfo(ci *container.Container) error {
	errFormat := "recordContainerInfo: %w"
	curPath := constants.CONTAINER_PATH + "/" + ci.Id
	container.MkDirErrorExit(curPath, 0755)
	bs, err := json.Marshal(ci)
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := os.WriteFile(curPath+"/config.json", bs, 0755); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func getContainerInfo(fs []string) []*container.Container {
	// errFormat := "getContainerInfo: %w"
	var cis []*container.Container

	for _, f := range fs {
		var c container.Container
		bs, err := os.ReadFile(f)
		if err != nil && err != io.EOF {
			log.Println("[error] read json file:", err)
			continue
		}
		if err := json.Unmarshal(bs, &c); err != nil {
			log.Println("[error] read json file:", err)
			continue
		}
		cis = append(cis, &c)
	}
	return cis
}

func printContainerInfo(ci []*container.Container) {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, err := fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	if err != nil {
		log.Println("[error] printContainerInfo:", err)
	}
	for _, c := range ci {
		_, err := fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			c.Id,
			c.Name,
			c.Pid,
			c.Status,
			c.Cmds,
			c.CreateAt,
		)
		if err != nil {
			log.Println("[error] printContainerInfo:", err)
		}
	}
	if err := w.Flush(); err != nil {
		log.Println("[error] printContainerInfo:", err)
	}
}

func findAllJsonFile(dir string) []string {
	var jsonfiles []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("[error] filepath.Walk:", path, err)
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			jsonfiles = append(jsonfiles, path)
		}
		return nil
	})
	return jsonfiles
}
