package container

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

const fdIndex = 3

func RunContainerInitProcess() error {
	errFormat := "runContainerInitProcess: %w"
	setupMount() // 必需先挂载，否者后续在LookPath会提示找不到路径
	// 从 pipe 读取命令
	cmdArray := readUserCommand()
	if len(cmdArray) == 0 {
		return errors.New("run container get user command error, cmdArray is nil")
	}
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		return fmt.Errorf(errFormat, err)
	}
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		return fmt.Errorf(errFormat, err)
	}
	return nil
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(fdIndex), "pipe")
	defer pipe.Close()
	msg, err := io.ReadAll(pipe)
	if err != nil {
		log.Println("[error] init read pipe:", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func setupMount() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Println("[error] os.Getwd:", err)
		return
	}
	log.Println("[debug] current location is", pwd)
	// systemd 加入linux之后, mount namespace 就变成 shared by default, 所以你必须显示声明你要这个新的mount namespace独立。
	// 即 mount proc 之前先把所有挂载点的传播类型改为 private，避免本 namespace 中的挂载事件外泄。
	if err := unix.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		log.Println("[error] setupMount syscall.Mount:", err)
	} // 测试执行这个操作也正常
	if err := pivotRout(pwd); err != nil {
		log.Println("[error] setupMount:", err)
		return
	}
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	unix.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	unix.Mount("tmpfs", "/dev", "tmpfs", unix.MS_NOSUID|unix.MS_STRICTATIME, "mode=755")
}

func pivotRout(root string) error {
	errFormat := "pivotRout %s: %w"
	if err := unix.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf(errFormat, "unix.Mount", err)
	}
	pivotDir := filepath.Join(root, ".pivot_root")
	log.Println("[debug] pivotDir is  " + pivotDir + ", when unix.PivotRoot")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return fmt.Errorf(errFormat, "os.Mkdir", err)
	}
	if err := unix.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf(errFormat, "unix.PivotRoot", err)
	}
	if err := unix.Chdir("/"); err != nil {
		return fmt.Errorf(errFormat, "unix.Chdir", err)
	}
	pivotDir = filepath.Join("/", ".pivot_root")
	log.Println("[debug] pivotDir is " + pivotDir + ", when unix.Unmount")
	if err := unix.Unmount(pivotDir, unix.MNT_DETACH); err != nil {
		return fmt.Errorf(errFormat, "unix.Unmout", err)
	}
	return os.Remove(pivotDir)
}
