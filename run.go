package main

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wlbyte/mydocker/container"
)

func Run(tty bool, comArray []string) {
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Printf("New parent error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	sendInitCommand(comArray, writePipe)
	_ = parent.Wait()
}

// sendInitCommand 通过writePipe将指令发送给子进程
func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Info("command all is ", command)
	_, _ = writePipe.WriteString(command)
	_ = writePipe.Close()
}
