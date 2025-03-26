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
	if err := unix.Exec(path, cmdArray[0:], os.Environ()); err != nil {
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
	// 在容器启动时，挂载命名空间默认是共享的，尤其是当系统使用systemd时。因此，需要显式地将根目录的传播类型设置为私有，
	// 以确保容器内的挂载操作不会影响到宿主机或其他容器。这一步通常在设置新的mount namespace之后执行，以确保隔离
	// 这行代码的作用是递归地将根目录及其子挂载点的传播类型设置为私有，为后续的挂载操作（如proc、tmpfs）提供隔离环境，避免影响宿主机或其他容器。
	// 即 mount proc 之前先把所有挂载点的传播类型改为 private，避免本 namespace 中的挂载事件外泄。
	if err := unix.Mount("", "/", "", unix.MS_PRIVATE|unix.MS_REC, ""); err != nil {
		log.Println("[error] setupMount unix.Mount:", err)
	}
	if err := pivotRout(pwd); err != nil {
		log.Println("[error] setupMount:", err)
		return
	}
	defaultMountFlags := unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV
	unix.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	unix.Mount("tmpfs", "/dev", "tmpfs", unix.MS_NOSUID|unix.MS_STRICTATIME, "mode=755")
}

func pivotRout(root string) error {
	errFormat := "pivotRout %s: %w"
	if err := unix.Mount(root, root, "bind", unix.MS_BIND|unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
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
