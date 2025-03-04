package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/wlbyte/mydocker/container"
)

func Run(tty bool, cmd string) {
	parent := container.NewParentProcess(tty, cmd)
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	_ = parent.Wait()
	os.Exit(1)
}
