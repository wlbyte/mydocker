package container

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const fdIndex = 3

func RunContainerInitProcess() error {
	mountProc()
	// 从 pipe 读取命令
	cmdArray := readUserCommand()
	if len(cmdArray) == 0 {
		return errors.New("run container get user command error, cmdArray is nil")
	}
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Printf("Exec loop path error %v", err)
		return err
	}
	log.Printf("Find path %s", path)

	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		log.Println(err.Error())
	}
	return nil
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(fdIndex), "pipe")
	defer pipe.Close()
	msg, err := io.ReadAll(pipe)
	if err != nil {
		log.Println("init read pipe error ", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func mountProc() {
	// systemd 加入linux之后, mount namespace 就变成 shared by default, 所以你必须显示声明你要这个新的mount namespace独立。
	// 即 mount proc 之前先把所有挂载点的传播类型改为 private，避免本 namespace 中的挂载事件外泄。
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "") // 测试执行这个操作也正常
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	_ = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
}
